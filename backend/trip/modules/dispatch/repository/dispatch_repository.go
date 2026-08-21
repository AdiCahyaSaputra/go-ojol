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
	PendingArgoTransaction(req dto.PendingArgoTransaction) error
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

func (r *dispatchRepository) PendingArgoTransaction(req dto.PendingArgoTransaction) error {
	data := entities.Transaction{
		CustomerID:         &req.CustomerID,
		VehicleID:          &req.VehicleID,
		PickupLatLong:      req.PickupLatLong[:],
		LastLatLong:        req.PickupLatLong[:],
		DestinationLatLong: req.DestinationLatLong[:],
		Distance:           req.Distance,
		FarePerDistance:    req.FarePerDistance,
		PlatformPercentage: req.PlatformPercentage,
		TotalFare:          req.TotalFare,
		Status:             entities.TransactionStatusPending,
	}

	result := r.db.Create(&data)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
