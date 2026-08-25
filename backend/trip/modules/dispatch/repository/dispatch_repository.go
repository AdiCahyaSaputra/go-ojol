package repository

import (
	"errors"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DispatchRepository interface {
	VehicleById(id uuid.UUID) (*entities.Vehicle, error)
	DistinctVehicleCategories() ([]dto.VehicleCategory, error)
	NearbyDriverProfiles(driverUserIds []uuid.UUID, vehicleType entities.VehicleType) (map[string]dto.NearbyDriverProfile, error)
	CreateOfferedTransaction(req dto.CreateOfferedTransaction) (*entities.Transaction, error)
	ClaimOffer(txID, driverID, vehicleID uuid.UUID) (bool, error)
	ExpireOffer(txID uuid.UUID) (bool, error)
	MarkRejectedOffer(txID uuid.UUID) (bool, error)
	TransactionByID(txID uuid.UUID) (*entities.Transaction, error)
}

type dispatchRepository struct {
	db *gorm.DB
}

func NewDispatchRepository(db *gorm.DB) DispatchRepository {
	return &dispatchRepository{
		db: db,
	}
}

func (r *dispatchRepository) VehicleById(id uuid.UUID) (*entities.Vehicle, error) {
	var vehicle *entities.Vehicle

	result := r.db.Find(&vehicle, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New(dto.MESSAGE_VEHICLE_NOT_FOUND)
		}
	}

	return vehicle, nil
}

func (r *dispatchRepository) DistinctVehicleCategories() ([]dto.VehicleCategory, error) {
	var categories []dto.VehicleCategory

	err := r.db.
		Table("vehicles").
		Select("distinct type, max_size").
		Order("type, max_size").
		Scan(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *dispatchRepository) NearbyDriverProfiles(driverUserIds []uuid.UUID, vehicleType entities.VehicleType) (map[string]dto.NearbyDriverProfile, error) {
	driversProfileMap := map[string]dto.NearbyDriverProfile{}
	var driversProfile []dto.NearbyDriverProfile

	err := r.db.
		Table("drivers d").
		Select(`
			d.id as driver_id,
			d.user_id,
			d.name,
			d.phone_number,
			d.profile_picture_url,
			v.id as vehicle_id,
			v.name as vehicle_name,
			v.license_number,
			v.max_size,
			v.type
		`).
		Joins(`
			join vehicles v
			on d.vehicle_id = v.id
		`).
		Where("v.type = ? and d.user_id in ?", vehicleType, driverUserIds).
		Scan(&driversProfile).Error

	if err != nil {
		return driversProfileMap, err
	}

	for _, driverProfile := range driversProfile {
		driversProfileMap[driverProfile.UserID.String()] = driverProfile
	}

	return driversProfileMap, nil
}

func (r *dispatchRepository) CreateOfferedTransaction(req dto.CreateOfferedTransaction) (*entities.Transaction, error) {
	data := entities.Transaction{
		CustomerID:         &req.CustomerID,
		PickupLatLong:      req.PickupLatLong[:],
		LastLatLong:        req.LastLatLong[:],
		DestinationLatLong: req.DestinationLatLong[:],
		Distance:           req.Distance,
		FarePerDistance:    req.FarePerDistance,
		PlatformPercentage: req.PlatformPercentage,
		TotalFare:          req.TotalFare,
		Status:             entities.TransactionStatusOffered,
	}

	if err := r.db.Create(&data).Error; err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *dispatchRepository) ClaimOffer(txID, driverID, vehicleID uuid.UUID) (bool, error) {
	result := r.db.Model(&entities.Transaction{}).
		Where("id = ? AND status = ?", txID, entities.TransactionStatusOffered).
		Updates(map[string]any{
			"status":     entities.TransactionStatusAcceptedOffer,
			"driver_id":  driverID,
			"vehicle_id": vehicleID,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *dispatchRepository) ExpireOffer(txID uuid.UUID) (bool, error) {
	result := r.db.Model(&entities.Transaction{}).
		Where("id = ? AND status = ?", txID, entities.TransactionStatusOffered).
		Update("status", entities.TransactionStatusExpired)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *dispatchRepository) MarkRejectedOffer(txID uuid.UUID) (bool, error) {
	result := r.db.Model(&entities.Transaction{}).
		Where("id = ? AND status = ?", txID, entities.TransactionStatusOffered).
		Update("status", entities.TransactionStatusRejectedOffer)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *dispatchRepository) TransactionByID(txID uuid.UUID) (*entities.Transaction, error) {
	var txn entities.Transaction
	if err := r.db.First(&txn, "id = ?", txID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(dto.MESSAGE_OFFER_NOT_FOUND)
		}
		return nil, err
	}
	return &txn, nil
}
