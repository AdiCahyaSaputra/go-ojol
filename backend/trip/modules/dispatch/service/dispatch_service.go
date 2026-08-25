package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/repository"
	wsdto "github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatchws/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/drivergeo"
	"gorm.io/gorm"
)

const (
	defaultOSRMBaseURL = "https://router.project-osrm.org"
	osrmUserAgent      = "go-ojol/trip"
	osrmTimeout        = 10 * time.Second

	farePerKmMotorcycle = 2500 // IDR
	farePerKmCar        = 4500 // IDR
	platformPct         = 10
	maxSizeFareStepPct  = 2
)

var (
	ErrNoRoute            = errors.New("no route found")
	ErrOSRMUnavailable    = errors.New("routing service unavailable")
	ErrInvalidLatLong     = errors.New("invalid lat long")
	ErrLocationStore      = errors.New("location store unavailable")
	ErrCustomerNotInCtx   = errors.New(dto.MESSAGE_CUSTOMER_NOT_FOUND_CTX)
	ErrDriverNotInCtx     = errors.New(dto.MESSAGE_DRIVER_NOT_FOUND_CTX)
	ErrOfferNotFound      = errors.New(dto.MESSAGE_OFFER_NOT_FOUND)
	ErrOfferUnavailable   = errors.New(dto.MESSAGE_OFFER_UNAVAILABLE)
	ErrInvalidOfferAction = errors.New(dto.MESSAGE_INVALID_OFFER_ACTION)
	ErrNotOfferedDriver   = errors.New(dto.MESSAGE_NOT_OFFERED_DRIVER)
	ErrNoLastSearch       = errors.New(dto.MESSAGE_NO_LAST_SEARCH)
	ErrOfferStillActive   = errors.New(dto.MESSAGE_OFFER_STILL_ACTIVE)
)

type OfferNotifier interface {
	Notify(userID string, msg wsdto.ServerMessage) bool
	NotifyMany(userIDs []string, msg wsdto.ServerMessage)
}

// DriverNotifier is kept as an alias for existing call sites/tests.
type DriverNotifier = OfferNotifier

type DispatchService interface {
	// Customer
	CalculateArgo(ctx context.Context, req dto.CalculateArgoRequest) (dto.CalculateArgoResponse, error)
	FindDriver(ctx context.Context, req dto.FindDriverRequest) (dto.FindDriverResponse, error)
	RetryFindDriver(ctx context.Context, customerUserID string) error

	// Driver
	SetDriverMode(ctx context.Context, req dto.SetDriverModeRequest) error
	RespondOffer(ctx context.Context, transactionID uuid.UUID, req dto.RespondOfferRequest) (dto.RespondOfferResponse, error)
}

type pendingOffer struct {
	customerUserID string
	driverUserIDs  []string
	pending        map[string]struct{}
}

type lastSearch struct {
	req      dto.FindDriverRequest
	customer entities.Customer
}

type dispatchService struct {
	dispatchRepository repository.DispatchRepository
	db                 *gorm.DB
	httpClient         *http.Client
	osrmBaseURL        string
	locations          *drivergeo.Store
	notifier           OfferNotifier
	offerTTL           time.Duration

	offersMu sync.Mutex
	offers   map[uuid.UUID]*pendingOffer

	searchesMu  sync.Mutex
	lastSearch  map[string]lastSearch
}

type Option interface {
	apply(*dispatchService)
}

type offerTTLOption time.Duration

func (o offerTTLOption) apply(s *dispatchService) {
	if time.Duration(o) > 0 {
		s.offerTTL = time.Duration(o)
	}
}

func WithOfferTTL(ttl time.Duration) Option {
	return offerTTLOption(ttl)
}

func NewDispatchService(
	dispatchRepo repository.DispatchRepository,
	db *gorm.DB,
	httpClient *http.Client,
	osrmBaseURL string,
	locations *drivergeo.Store,
	notifier OfferNotifier,
	opts ...Option,
) DispatchService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: osrmTimeout}
	}
	if osrmBaseURL == "" {
		osrmBaseURL = defaultOSRMBaseURL
	}

	s := &dispatchService{
		dispatchRepository: dispatchRepo,
		db:                 db,
		httpClient:         httpClient,
		osrmBaseURL:        strings.TrimRight(osrmBaseURL, "/"),
		locations:          locations,
		notifier:           notifier,
		offerTTL:           time.Duration(dto.OfferTTLSeconds) * time.Second,
		offers:             make(map[uuid.UUID]*pendingOffer),
		lastSearch:         make(map[string]lastSearch),
	}
	for _, opt := range opts {
		opt.apply(s)
	}
	return s
}

func (s *dispatchService) CalculateArgo(ctx context.Context, req dto.CalculateArgoRequest) (dto.CalculateArgoResponse, error) {
	pickupLat, pickupLng, err := parseLatLong(req.PickupLoc)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}
	destLat, destLng, err := parseLatLong(req.Destination)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	route, err := s.getOSRMRoute(ctx, pickupLat, pickupLng, destLat, destLng)
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	categories, err := s.dispatchRepository.DistinctVehicleCategories()
	if err != nil {
		return dto.CalculateArgoResponse{}, err
	}

	distance := int(math.Round(route.Distance))
	duration := int(math.Round(route.Duration))
	vehicleOptions := make([]dto.VehicleOption, 0, len(categories))

	for _, category := range categories {
		farePerDistance, ok := farePerDistanceFor(category.VehicleType)
		if !ok {
			continue
		}

		vehicleOptions = append(vehicleOptions, dto.VehicleOption{
			VehicleType: category.VehicleType,
			MaxSize:     category.MaxSize,
			TotalFare:   calculateTotalFare(distance, farePerDistance, category.MaxSize),
		})
	}

	return dto.CalculateArgoResponse{
		Distance:           distance,
		Duration:           duration,
		Path:               geoJSONToLatLngPath(route.Geometry.Coordinates),
		PlatformPercentage: platformPct,
		VehicleOptions:     vehicleOptions,
	}, nil
}

func (s *dispatchService) FindDriver(ctx context.Context, req dto.FindDriverRequest) (dto.FindDriverResponse, error) {
	customer, ok := ctx.Value("customer").(entities.Customer)
	if !ok || customer.ID == uuid.Nil {
		return dto.FindDriverResponse{}, ErrCustomerNotInCtx
	}

	return s.startOffer(ctx, customer, req)
}

func (s *dispatchService) RetryFindDriver(ctx context.Context, customerUserID string) error {
	if customerUserID == "" {
		return ErrNoLastSearch
	}

	s.searchesMu.Lock()
	search, ok := s.lastSearch[customerUserID]
	s.searchesMu.Unlock()
	if !ok {
		return ErrNoLastSearch
	}

	if s.hasActiveOfferForCustomer(customerUserID) {
		return ErrOfferStillActive
	}

	_, err := s.startOffer(ctx, search.customer, search.req)
	return err
}

func (s *dispatchService) startOffer(ctx context.Context, customer entities.Customer, req dto.FindDriverRequest) (dto.FindDriverResponse, error) {
	if s.locations == nil {
		return dto.FindDriverResponse{}, ErrLocationStore
	}

	customerUserID := customer.UserID.String()
	if customer.UserID == uuid.Nil {
		return dto.FindDriverResponse{}, ErrCustomerNotInCtx
	}

	pickupLat, pickupLng, err := parseLatLong(req.PickupLatLong)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}
	destLat, destLng, err := parseLatLong(req.DestinationLatLong)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	nearby, err := s.locations.Nearby(ctx, pickupLat, pickupLng, drivergeo.DefaultRadiusKm, drivergeo.DefaultCount)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	driverUserIds := make([]uuid.UUID, 0, len(nearby))
	for _, driver := range nearby {
		driverUserIds = append(driverUserIds, uuid.MustParse(driver.UserID))
	}

	nearbyDriverProfiles, err := s.dispatchRepository.NearbyDriverProfiles(driverUserIds, req.VehicleType)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	drivers := make([]dto.NearbyDriver, 0, len(nearby))
	for _, driver := range nearby {
		if driverProfile, ok := nearbyDriverProfiles[driver.UserID]; ok {
			if driverProfile.MaxSize < req.MaxSize {
				continue
			}
			drivers = append(drivers, dto.NearbyDriver{
				UserID:    driver.UserID,
				DistanceM: driver.DistanceM,
				Location:  [2]float64{driver.Lat, driver.Lng},
				Profile:   driverProfile,
			})
		}
	}

	s.storeLastSearch(customerUserID, customer, req)

	if len(drivers) == 0 {
		if s.notifier != nil {
			s.notifier.Notify(customerUserID, wsdto.ServerMessage{Type: wsdto.TypeNoDrivers})
		}
		return dto.FindDriverResponse{Drivers: drivers}, nil
	}

	route, err := s.getOSRMRoute(ctx, pickupLat, pickupLng, destLat, destLng)
	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	farePerDistance, ok := farePerDistanceFor(req.VehicleType)
	if !ok {
		return dto.FindDriverResponse{}, ErrInvalidLatLong
	}

	distance := int(math.Round(route.Distance))
	totalFare := calculateTotalFare(distance, farePerDistance, req.MaxSize)

	transactionNew, err := s.dispatchRepository.CreateOfferedTransaction(dto.CreateOfferedTransaction{
		CustomerID:         customer.ID,
		PickupLatLong:      req.PickupLatLong,
		DestinationLatLong: req.DestinationLatLong,
		LastLatLong:        req.PickupLatLong,
		Distance:           distance,
		FarePerDistance:    farePerDistance,
		PlatformPercentage: platformPct,
		TotalFare:          totalFare,
	})

	if err != nil {
		return dto.FindDriverResponse{}, err
	}

	offeredUserIDs := make([]string, 0, len(drivers))
	pending := make(map[string]struct{}, len(drivers))
	for _, d := range drivers {
		offeredUserIDs = append(offeredUserIDs, d.UserID)
		pending[d.UserID] = struct{}{}
	}

	s.offersMu.Lock()
	s.offers[transactionNew.ID] = &pendingOffer{
		customerUserID: customerUserID,
		driverUserIDs:  offeredUserIDs,
		pending:        pending,
	}
	s.offersMu.Unlock()

	offerMsg := wsdto.ServerMessage{
		Type:          wsdto.TypeTripOffer,
		TransactionID: transactionNew.ID.String(),
		Offer: &wsdto.TripOfferPayload{
			TransactionID: transactionNew.ID.String(),
			CustomerName:  customer.Name,
			Pickup:        [2]float64{pickupLat, pickupLng},
			Destination:   [2]float64{destLat, destLng},
			DistanceM:     distance,
			TotalFare:     totalFare,
			ExpiresInSec:  dto.OfferTTLSeconds,
		},
	}
	if s.notifier != nil {
		s.notifier.NotifyMany(offeredUserIDs, offerMsg)
		s.notifier.Notify(customerUserID, wsdto.ServerMessage{
			Type:          wsdto.TypeWaiting,
			TransactionID: transactionNew.ID.String(),
			ExpiresInSec:  dto.OfferTTLSeconds,
		})
	}

	s.scheduleOfferExpiry(transactionNew.ID)

	return dto.FindDriverResponse{
		TransactionID: &transactionNew.ID,
		ExpiresInSec:  dto.OfferTTLSeconds,
		Drivers:       drivers,
	}, nil
}

func (s *dispatchService) storeLastSearch(customerUserID string, customer entities.Customer, req dto.FindDriverRequest) {
	s.searchesMu.Lock()
	defer s.searchesMu.Unlock()
	s.lastSearch[customerUserID] = lastSearch{req: req, customer: customer}
}

func (s *dispatchService) hasActiveOfferForCustomer(customerUserID string) bool {
	s.offersMu.Lock()
	defer s.offersMu.Unlock()
	for _, offer := range s.offers {
		if offer.customerUserID == customerUserID {
			return true
		}
	}
	return false
}

func (s *dispatchService) scheduleOfferExpiry(txID uuid.UUID) {
	ttl := s.offerTTL
	time.AfterFunc(ttl, func() {
		ok, err := s.dispatchRepository.ExpireOffer(txID)
		if err != nil || !ok {
			s.clearPendingOffer(txID)
			return
		}

		offer := s.clearPendingOffer(txID)
		if offer == nil || s.notifier == nil {
			return
		}
		if len(offer.driverUserIDs) > 0 {
			s.notifier.NotifyMany(offer.driverUserIDs, wsdto.ServerMessage{
				Type:          wsdto.TypeOfferExpired,
				TransactionID: txID.String(),
			})
		}
		if offer.customerUserID != "" {
			s.notifier.Notify(offer.customerUserID, wsdto.ServerMessage{
				Type:          wsdto.TypeOfferExpired,
				TransactionID: txID.String(),
			})
		}
	})
}

func (s *dispatchService) clearPendingOffer(txID uuid.UUID) *pendingOffer {
	s.offersMu.Lock()
	defer s.offersMu.Unlock()
	offer, ok := s.offers[txID]
	if !ok {
		return nil
	}
	delete(s.offers, txID)
	return offer
}

func (s *dispatchService) RespondOffer(ctx context.Context, transactionID uuid.UUID, req dto.RespondOfferRequest) (dto.RespondOfferResponse, error) {
	driver, ok := ctx.Value("driver").(entities.Driver)
	if !ok || driver.ID == uuid.Nil {
		return dto.RespondOfferResponse{}, ErrDriverNotInCtx
	}

	switch req.Action {
	case dto.OfferActionAccept:
		return s.acceptOffer(ctx, transactionID, driver)
	case dto.OfferActionReject:
		return s.rejectOffer(transactionID, driver)
	default:
		return dto.RespondOfferResponse{}, ErrInvalidOfferAction
	}
}

func (s *dispatchService) acceptOffer(ctx context.Context, transactionID uuid.UUID, driver entities.Driver) (dto.RespondOfferResponse, error) {
	if !s.isOfferedDriver(transactionID, driver.UserID.String()) {
		txn, err := s.dispatchRepository.TransactionByID(transactionID)
		if err != nil {
			return dto.RespondOfferResponse{}, mapOfferLookupErr(err)
		}
		if txn.Status != entities.TransactionStatusOffered {
			return dto.RespondOfferResponse{}, ErrOfferUnavailable
		}
		return dto.RespondOfferResponse{}, ErrNotOfferedDriver
	}

	claimed, err := s.dispatchRepository.ClaimOffer(transactionID, driver.ID, driver.VehicleID)
	if err != nil {
		return dto.RespondOfferResponse{}, err
	}
	if !claimed {
		return dto.RespondOfferResponse{}, ErrOfferUnavailable
	}

	if s.locations != nil {
		_ = s.locations.RemoveStandby(ctx, driver.UserID.String())
	}

	offer := s.clearPendingOffer(transactionID)
	others := make([]string, 0)
	if offer != nil {
		winner := driver.UserID.String()
		for _, id := range offer.driverUserIDs {
			if id != winner {
				others = append(others, id)
			}
		}
	}
	if s.notifier != nil && len(others) > 0 {
		s.notifier.NotifyMany(others, wsdto.ServerMessage{
			Type:          wsdto.TypeOfferTaken,
			TransactionID: transactionID.String(),
		})
	}

	if s.notifier != nil && offer != nil && offer.customerUserID != "" {
		matched := &wsdto.MatchedDriverPayload{
			UserID:      driver.UserID.String(),
			DriverID:    driver.ID.String(),
			Name:        driver.Name,
			PhoneNumber: driver.PhoneNumber,
			VehicleID:   driver.VehicleID.String(),
		}
		if vehicle, err := s.dispatchRepository.VehicleById(driver.VehicleID); err == nil && vehicle != nil {
			matched.VehicleName = vehicle.Name
			matched.LicenseNumber = vehicle.LicenseNumber
			matched.VehicleType = string(vehicle.Type)
		}
		s.notifier.Notify(offer.customerUserID, wsdto.ServerMessage{
			Type:          wsdto.TypeDriverMatched,
			TransactionID: transactionID.String(),
			MatchedDriver: matched,
		})
	}

	return dto.RespondOfferResponse{
		TransactionID: transactionID,
		Status:        entities.TransactionStatusAcceptedOffer,
	}, nil
}

func (s *dispatchService) rejectOffer(transactionID uuid.UUID, driver entities.Driver) (dto.RespondOfferResponse, error) {
	s.offersMu.Lock()
	offer, ok := s.offers[transactionID]
	if !ok {
		s.offersMu.Unlock()
		txn, err := s.dispatchRepository.TransactionByID(transactionID)
		if err != nil {
			return dto.RespondOfferResponse{}, mapOfferLookupErr(err)
		}
		if txn.Status != entities.TransactionStatusOffered {
			return dto.RespondOfferResponse{}, ErrOfferUnavailable
		}
		return dto.RespondOfferResponse{}, ErrNotOfferedDriver
	}

	userID := driver.UserID.String()
	if _, offered := offer.pending[userID]; !offered {
		s.offersMu.Unlock()
		return dto.RespondOfferResponse{}, ErrNotOfferedDriver
	}
	delete(offer.pending, userID)
	allRejected := len(offer.pending) == 0
	customerUserID := offer.customerUserID
	s.offersMu.Unlock()

	status := entities.TransactionStatusOffered
	if allRejected {
		ok, err := s.dispatchRepository.MarkRejectedOffer(transactionID)
		if err != nil {
			return dto.RespondOfferResponse{}, err
		}
		if ok {
			status = entities.TransactionStatusRejectedOffer
		}
		s.clearPendingOffer(transactionID)
		if s.notifier != nil && customerUserID != "" {
			s.notifier.Notify(customerUserID, wsdto.ServerMessage{
				Type:          wsdto.TypeOfferRejected,
				TransactionID: transactionID.String(),
			})
		}
	}

	return dto.RespondOfferResponse{
		TransactionID: transactionID,
		Status:        status,
	}, nil
}

func (s *dispatchService) isOfferedDriver(txID uuid.UUID, userID string) bool {
	s.offersMu.Lock()
	defer s.offersMu.Unlock()
	offer, ok := s.offers[txID]
	if !ok {
		return false
	}
	_, pending := offer.pending[userID]
	return pending
}

func mapOfferLookupErr(err error) error {
	if err != nil && err.Error() == dto.MESSAGE_OFFER_NOT_FOUND {
		return ErrOfferNotFound
	}
	return err
}

type osrmRouteResponse struct {
	Code   string      `json:"code"`
	Routes []osrmRoute `json:"routes"`
}

type osrmRoute struct {
	Distance float64      `json:"distance"`
	Duration float64      `json:"duration"`
	Geometry osrmGeometry `json:"geometry"`
}

type osrmGeometry struct {
	Coordinates [][]float64 `json:"coordinates"`
}

func (s *dispatchService) getOSRMRoute(ctx context.Context, pickupLat, pickupLng, destLat, destLng float64) (osrmRoute, error) {
	url := fmt.Sprintf(
		"%s/route/v1/driving/%s,%s;%s,%s?geometries=geojson&overview=full",
		s.osrmBaseURL,
		formatCoord(pickupLng), // Notice that we pass longitude first
		formatCoord(pickupLat),
		formatCoord(destLng),
		formatCoord(destLat),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return osrmRoute{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}
	req.Header.Set("User-Agent", osrmUserAgent)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return osrmRoute{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return osrmRoute{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}

	if res.StatusCode >= 500 {
		return osrmRoute{}, fmt.Errorf("%w: status %d", ErrOSRMUnavailable, res.StatusCode)
	}
	if res.StatusCode >= 400 {
		return osrmRoute{}, ErrNoRoute
	}

	var parsed osrmRouteResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return osrmRoute{}, fmt.Errorf("%w: %v", ErrOSRMUnavailable, err)
	}

	if parsed.Code != "Ok" || len(parsed.Routes) == 0 {
		return osrmRoute{}, ErrNoRoute
	}

	return parsed.Routes[0], nil
}

func (s *dispatchService) SetDriverMode(ctx context.Context, req dto.SetDriverModeRequest) error {
	userId := ctx.Value("user_id").(string)

	if userId == "" {
		return errors.New(dto.MESSAGE_DRIVE_USER_ID_CONTEXT_NOT_FOUND)
	}

	lat, long, err := parseLatLong(req.CurrentLatLong)

	if err != nil {
		return err
	}

	switch req.Mode {
	case dto.DriverModeOnline:
		err = s.locations.SetStandby(ctx, userId, lat, long)

		if err != nil {
			return err
		}
	case dto.DriverModeOffline:
		err = s.locations.RemoveStandby(ctx, userId)

		if err != nil {
			return err
		}
	}

	return nil
}

func parseLatLong(pair [2]string) (lat float64, lng float64, err error) {
	lat, errLat := strconv.ParseFloat(pair[0], 64)
	lng, errLng := strconv.ParseFloat(pair[1], 64)
	if errLat != nil || errLng != nil {
		return 0, 0, ErrInvalidLatLong
	}
	return lat, lng, nil
}

func formatCoord(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func farePerDistanceFor(vehicleType entities.VehicleType) (int, bool) {
	switch vehicleType {
	case entities.VehicleTypeMotorcycle:
		return farePerKmMotorcycle, true
	case entities.VehicleTypeCar:
		return farePerKmCar, true
	default:
		return 0, false
	}
}

func calculateTotalFare(distance, farePerDistance, maxSize int) int {
	sizePct := 100
	if maxSize > 1 {
		sizePct += (maxSize - 1) * maxSizeFareStepPct
	}

	numerator := int64(distance) * int64(farePerDistance) * int64(sizePct) * int64(100+platformPct)
	denominator := int64(1000 * 100 * 100)

	return int((numerator + denominator - 1) / denominator)
}

// The geoJSON response is in reverse [long, lat] so we need to re-reverse it
func geoJSONToLatLngPath(coordinates [][]float64) [][2]float64 {
	path := make([][2]float64, 0, len(coordinates))
	for _, coord := range coordinates {
		if len(coord) < 2 {
			continue
		}
		path = append(path, [2]float64{coord[1], coord[0]})
	}
	return path
}
