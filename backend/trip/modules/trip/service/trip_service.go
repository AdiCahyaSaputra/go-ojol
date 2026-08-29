package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/repository"
	wsdto "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/triploc"
	"github.com/google/uuid"
)

var (
	ErrTransactionNotFound      = errors.New(dto.MESSAGE_TRANSACTION_NOT_FOUND)
	ErrNoActiveTransaction      = errors.New(dto.MESSAGE_NO_ACTIVE_TRANSACTION)
	ErrNotTransactionParticipant = errors.New(dto.MESSAGE_NOT_TRANSACTION_PARTICIPANT)
	ErrInvalidStatusTransition  = errors.New(dto.MESSAGE_INVALID_STATUS_TRANSITION)
)

type TripNotifier interface {
	Notify(userID string, msg wsdto.ServerMessage) bool
}

type TripService interface {
	GetActiveForCustomer(ctx context.Context, customer entities.Customer) (dto.TransactionResponse, error)
	GetActiveForDriver(ctx context.Context, driver entities.Driver) (dto.TransactionResponse, error)
	GetByIDForCustomer(ctx context.Context, customer entities.Customer, txID uuid.UUID) (dto.TransactionResponse, error)
	GetByIDForDriver(ctx context.Context, driver entities.Driver, txID uuid.UUID) (dto.TransactionResponse, error)
	StartTrip(ctx context.Context, driver entities.Driver, txID uuid.UUID) (dto.StartTripResponse, error)
	CompleteTrip(ctx context.Context, driver entities.Driver, txID uuid.UUID) (dto.CompleteTripResponse, error)
	CancelTripAsCustomer(ctx context.Context, customer entities.Customer, txID uuid.UUID) (dto.CancelTripResponse, error)
	CancelTripAsDriver(ctx context.Context, driver entities.Driver, txID uuid.UUID) (dto.CancelTripResponse, error)
	HandleDriverTripLocation(ctx context.Context, driverUserID string, txID uuid.UUID, lat, lng float64) error
	HandleCustomerTripLocation(ctx context.Context, customerUserID string, txID uuid.UUID, lat, lng float64) error
}

type tripService struct {
	repo     repository.TripRepository
	locations *triploc.Store
	notifier TripNotifier
}

func NewTripService(repo repository.TripRepository, locations *triploc.Store, notifier TripNotifier) TripService {
	return &tripService{
		repo:      repo,
		locations: locations,
		notifier:  notifier,
	}
}

func (s *tripService) GetActiveForCustomer(ctx context.Context, customer entities.Customer) (dto.TransactionResponse, error) {
	txn, err := s.repo.ActiveByCustomerID(customer.ID)
	if err != nil {
		return dto.TransactionResponse{}, mapTripRepoErr(err)
	}
	return s.buildResponse(ctx, txn), nil
}

func (s *tripService) GetActiveForDriver(ctx context.Context, driver entities.Driver) (dto.TransactionResponse, error) {
	txn, err := s.repo.ActiveByDriverID(driver.ID)
	if err != nil {
		return dto.TransactionResponse{}, mapTripRepoErr(err)
	}
	return s.buildResponse(ctx, txn), nil
}

func (s *tripService) GetByIDForCustomer(ctx context.Context, customer entities.Customer, txID uuid.UUID) (dto.TransactionResponse, error) {
	txn, err := s.repo.TransactionWithRelations(txID)
	if err != nil {
		return dto.TransactionResponse{}, mapTripRepoErr(err)
	}
	if txn.CustomerID == nil || *txn.CustomerID != customer.ID {
		return dto.TransactionResponse{}, ErrNotTransactionParticipant
	}
	return s.buildResponse(ctx, txn), nil
}

func (s *tripService) GetByIDForDriver(ctx context.Context, driver entities.Driver, txID uuid.UUID) (dto.TransactionResponse, error) {
	txn, err := s.repo.TransactionWithRelations(txID)
	if err != nil {
		return dto.TransactionResponse{}, mapTripRepoErr(err)
	}
	if txn.DriverID == nil || *txn.DriverID != driver.ID {
		return dto.TransactionResponse{}, ErrNotTransactionParticipant
	}
	return s.buildResponse(ctx, txn), nil
}

func (s *tripService) StartTrip(ctx context.Context, driver entities.Driver, txID uuid.UUID) (dto.StartTripResponse, error) {
	txn, err := s.repo.TransactionWithRelations(txID)
	if err != nil {
		return dto.StartTripResponse{}, mapTripRepoErr(err)
	}
	if txn.DriverID == nil || *txn.DriverID != driver.ID {
		return dto.StartTripResponse{}, ErrNotTransactionParticipant
	}
	if txn.Status == entities.TransactionStatusOnTheWay {
		return dto.StartTripResponse{
			TransactionID: txID.String(),
			Status:        string(entities.TransactionStatusOnTheWay),
		}, nil
	}
	if txn.Status != entities.TransactionStatusAcceptedOffer {
		return dto.StartTripResponse{}, ErrInvalidStatusTransition
	}

	ok, err := s.repo.UpdateStatus(txID, entities.TransactionStatusAcceptedOffer, entities.TransactionStatusOnTheWay, nil)
	if err != nil {
		return dto.StartTripResponse{}, err
	}
	if !ok {
		return dto.StartTripResponse{}, ErrInvalidStatusTransition
	}

	s.broadcastTripStatus(txn, entities.TransactionStatusOnTheWay)

	return dto.StartTripResponse{
		TransactionID: txID.String(),
		Status:        string(entities.TransactionStatusOnTheWay),
	}, nil
}

func (s *tripService) CompleteTrip(ctx context.Context, driver entities.Driver, txID uuid.UUID) (dto.CompleteTripResponse, error) {
	txn, err := s.repo.TransactionWithRelations(txID)
	if err != nil {
		return dto.CompleteTripResponse{}, mapTripRepoErr(err)
	}
	if txn.DriverID == nil || *txn.DriverID != driver.ID {
		return dto.CompleteTripResponse{}, ErrNotTransactionParticipant
	}
	if txn.Status == entities.TransactionStatusCompleted {
		paidAt := ""
		if txn.PaidAt != nil {
			paidAt = txn.PaidAt.UTC().Format(time.RFC3339)
		}
		return dto.CompleteTripResponse{
			TransactionID: txID.String(),
			Status:        string(entities.TransactionStatusCompleted),
			TotalFare:     txn.TotalFare,
			PaidAt:        paidAt,
		}, nil
	}
	if txn.Status != entities.TransactionStatusOnTheWay {
		return dto.CompleteTripResponse{}, ErrInvalidStatusTransition
	}

	paidAt := time.Now().UTC()
	ok, err := s.repo.UpdateStatus(
		txID,
		entities.TransactionStatusOnTheWay,
		entities.TransactionStatusCompleted,
		repository.CompleteExtra(paidAt),
	)
	if err != nil {
		return dto.CompleteTripResponse{}, err
	}
	if !ok {
		return dto.CompleteTripResponse{}, ErrInvalidStatusTransition
	}

	if s.locations != nil {
		_ = s.locations.ClearTrip(ctx, txID.String())
	}

	txn.Status = entities.TransactionStatusCompleted
	txn.PaidAt = &paidAt
	s.broadcastTripCompleted(txn, paidAt)

	return dto.CompleteTripResponse{
		TransactionID: txID.String(),
		Status:        string(entities.TransactionStatusCompleted),
		TotalFare:     txn.TotalFare,
		PaidAt:        paidAt.Format(time.RFC3339),
	}, nil
}

func (s *tripService) CancelTripAsCustomer(ctx context.Context, customer entities.Customer, txID uuid.UUID) (dto.CancelTripResponse, error) {
	return s.cancelTrip(ctx, txID, func(txn *entities.Transaction) error {
		if txn.CustomerID == nil || *txn.CustomerID != customer.ID {
			return ErrNotTransactionParticipant
		}
		return nil
	})
}

func (s *tripService) CancelTripAsDriver(ctx context.Context, driver entities.Driver, txID uuid.UUID) (dto.CancelTripResponse, error) {
	return s.cancelTrip(ctx, txID, func(txn *entities.Transaction) error {
		if txn.DriverID == nil || *txn.DriverID != driver.ID {
			return ErrNotTransactionParticipant
		}
		return nil
	})
}

func (s *tripService) cancelTrip(ctx context.Context, txID uuid.UUID, authorize func(*entities.Transaction) error) (dto.CancelTripResponse, error) {
	txn, err := s.repo.TransactionWithRelations(txID)
	if err != nil {
		return dto.CancelTripResponse{}, mapTripRepoErr(err)
	}
	if err := authorize(txn); err != nil {
		return dto.CancelTripResponse{}, err
	}
	if txn.Status == entities.TransactionStatusCancelled {
		return dto.CancelTripResponse{
			TransactionID: txID.String(),
			Status:        string(entities.TransactionStatusCancelled),
		}, nil
	}
	if txn.Status != entities.TransactionStatusAcceptedOffer && txn.Status != entities.TransactionStatusOnTheWay {
		return dto.CancelTripResponse{}, ErrInvalidStatusTransition
	}

	from := txn.Status
	ok, err := s.repo.UpdateStatus(txID, from, entities.TransactionStatusCancelled, nil)
	if err != nil {
		return dto.CancelTripResponse{}, err
	}
	if !ok {
		return dto.CancelTripResponse{}, ErrInvalidStatusTransition
	}

	if s.locations != nil {
		_ = s.locations.ClearTrip(ctx, txID.String())
	}

	s.broadcastTripStatus(txn, entities.TransactionStatusCancelled)

	return dto.CancelTripResponse{
		TransactionID: txID.String(),
		Status:        string(entities.TransactionStatusCancelled),
	}, nil
}

func (s *tripService) HandleDriverTripLocation(ctx context.Context, driverUserID string, txID uuid.UUID, lat, lng float64) error {
	txn, err := s.repo.TransactionWithRelations(txID)
	if err != nil {
		return mapTripRepoErr(err)
	}
	if !isActiveTrip(txn) {
		return ErrInvalidStatusTransition
	}
	if txn.Driver == nil || txn.Driver.UserID.String() != driverUserID {
		return ErrNotTransactionParticipant
	}

	latStr := formatCoord(lat)
	lngStr := formatCoord(lng)
	if err := s.repo.UpdateDriverLastLatLong(txID, latStr, lngStr); err != nil {
		return err
	}
	if s.locations != nil {
		_ = s.locations.SetDriver(ctx, txID.String(), lat, lng)
	}

	if s.notifier != nil && txn.Customer != nil {
		s.notifier.Notify(txn.Customer.UserID.String(), wsdto.ServerMessage{
			Type:          wsdto.TypeDriverLocation,
			TransactionID: txID.String(),
			Lat:           lat,
			Lng:           lng,
		})
	}
	return nil
}

func (s *tripService) HandleCustomerTripLocation(ctx context.Context, customerUserID string, txID uuid.UUID, lat, lng float64) error {
	txn, err := s.repo.TransactionWithRelations(txID)
	if err != nil {
		return mapTripRepoErr(err)
	}
	if !isActiveTrip(txn) {
		return ErrInvalidStatusTransition
	}
	if txn.Customer == nil || txn.Customer.UserID.String() != customerUserID {
		return ErrNotTransactionParticipant
	}

	latStr := formatCoord(lat)
	lngStr := formatCoord(lng)
	if err := s.repo.UpdateCustomerLastLatLong(txID, latStr, lngStr); err != nil {
		return err
	}
	if s.locations != nil {
		_ = s.locations.SetCustomer(ctx, txID.String(), lat, lng)
	}

	if s.notifier != nil && txn.Driver != nil {
		s.notifier.Notify(txn.Driver.UserID.String(), wsdto.ServerMessage{
			Type:          wsdto.TypeCustomerLocation,
			TransactionID: txID.String(),
			Lat:           lat,
			Lng:           lng,
		})
	}
	return nil
}

func (s *tripService) buildResponse(ctx context.Context, txn *entities.Transaction) dto.TransactionResponse {
	resp := dto.TransactionResponse{
		ID:                 txn.ID.String(),
		Status:             string(txn.Status),
		PickupLatLong:      toLatLongPair(txn.PickupLatLong),
		DestinationLatLong: toLatLongPair(txn.DestinationLatLong),
		DriverLastLatLong:  toLatLongPair(txn.DriverLastLatLong),
		Distance:           txn.Distance,
		TotalFare:          txn.TotalFare,
	}

	if len(txn.CustomerLastLatLong) >= 2 {
		pair := toLatLongPair(txn.CustomerLastLatLong)
		resp.CustomerLastLatLong = &pair
	}

	if txn.PaidAt != nil {
		paidAt := txn.PaidAt.UTC().Format(time.RFC3339)
		resp.PaidAt = &paidAt
	}

	if s.locations != nil {
		if coords, ok, err := s.locations.GetDriver(ctx, txn.ID.String()); err == nil && ok {
			resp.DriverLastLatLong = dto.LatLongPair{coords.Lat, coords.Lng}
		}
		if coords, ok, err := s.locations.GetCustomer(ctx, txn.ID.String()); err == nil && ok {
			pair := dto.LatLongPair{coords.Lat, coords.Lng}
			resp.CustomerLastLatLong = &pair
		}
	}

	if txn.Driver != nil && txn.Vehicle != nil {
		resp.Driver = &dto.TripDriverProfile{
			UserID:        txn.Driver.UserID.String(),
			DriverID:      txn.Driver.ID.String(),
			Name:          txn.Driver.Name,
			PhoneNumber:   txn.Driver.PhoneNumber,
			VehicleID:     txn.Vehicle.ID.String(),
			VehicleName:   txn.Vehicle.Name,
			LicenseNumber: txn.Vehicle.LicenseNumber,
			VehicleType:   string(txn.Vehicle.Type),
		}
	}

	return resp
}

func (s *tripService) broadcastTripStatus(txn *entities.Transaction, status entities.TransactionStatus) {
	if s.notifier == nil {
		return
	}
	msg := wsdto.ServerMessage{
		Type:          wsdto.TypeTripStatus,
		TransactionID: txn.ID.String(),
		Status:        string(status),
	}
	if txn.Customer != nil {
		s.notifier.Notify(txn.Customer.UserID.String(), msg)
	}
	if txn.Driver != nil {
		s.notifier.Notify(txn.Driver.UserID.String(), msg)
	}
}

func (s *tripService) broadcastTripCompleted(txn *entities.Transaction, paidAt time.Time) {
	if s.notifier == nil {
		return
	}
	msg := wsdto.ServerMessage{
		Type:          wsdto.TypeTripCompleted,
		TransactionID: txn.ID.String(),
		Status:        string(entities.TransactionStatusCompleted),
		TotalFare:     txn.TotalFare,
		PaidAt:        paidAt.UTC().Format(time.RFC3339),
	}
	if txn.Customer != nil {
		s.notifier.Notify(txn.Customer.UserID.String(), msg)
	}
	if txn.Driver != nil {
		s.notifier.Notify(txn.Driver.UserID.String(), msg)
	}
}

func isActiveTrip(txn *entities.Transaction) bool {
	return txn.Status == entities.TransactionStatusAcceptedOffer || txn.Status == entities.TransactionStatusOnTheWay
}

func toLatLongPair(values []string) dto.LatLongPair {
	if len(values) < 2 {
		return dto.LatLongPair{0, 0}
	}
	lat, _ := strconv.ParseFloat(values[0], 64)
	lng, _ := strconv.ParseFloat(values[1], 64)
	return dto.LatLongPair{lat, lng}
}

func formatCoord(value float64) string {
	return fmt.Sprintf("%.8f", value)
}

func mapTripRepoErr(err error) error {
	if err == nil {
		return nil
	}
	switch err.Error() {
	case dto.MESSAGE_TRANSACTION_NOT_FOUND:
		return ErrTransactionNotFound
	case dto.MESSAGE_NO_ACTIVE_TRANSACTION:
		return ErrNoActiveTransaction
	default:
		return err
	}
}
