package repository

import (
	"maps"
	"errors"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/trip/dto"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var activeStatuses = []entities.TransactionStatus{
	entities.TransactionStatusAcceptedOffer,
	entities.TransactionStatusOnTheWay,
}

type TripRepository interface {
	ActiveByCustomerID(customerID uuid.UUID) (*entities.Transaction, error)
	ActiveByDriverID(driverID uuid.UUID) (*entities.Transaction, error)
	TransactionByID(txID uuid.UUID) (*entities.Transaction, error)
	TransactionWithRelations(txID uuid.UUID) (*entities.Transaction, error)
	UpdateStatus(txID uuid.UUID, from, to entities.TransactionStatus, extra map[string]any) (bool, error)
	UpdateDriverLastLatLong(txID uuid.UUID, lat, lng string) error
	UpdateCustomerLastLatLong(txID uuid.UUID, lat, lng string) error
}

type tripRepository struct {
	db *gorm.DB
}

func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}

func (r *tripRepository) ActiveByCustomerID(customerID uuid.UUID) (*entities.Transaction, error) {
	var txn entities.Transaction
	err := r.db.
		Preload("Driver").
		Preload("Vehicle").
		Where("customer_id = ? AND status IN ?", customerID, activeStatuses).
		Order("updated_at DESC").
		First(&txn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(dto.MESSAGE_NO_ACTIVE_TRANSACTION)
		}
		return nil, err
	}
	return &txn, nil
}

func (r *tripRepository) ActiveByDriverID(driverID uuid.UUID) (*entities.Transaction, error) {
	var txn entities.Transaction
	err := r.db.
		Preload("Customer").
		Where("driver_id = ? AND status IN ?", driverID, activeStatuses).
		Order("updated_at DESC").
		First(&txn).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(dto.MESSAGE_NO_ACTIVE_TRANSACTION)
		}
		return nil, err
	}
	return &txn, nil
}

func (r *tripRepository) TransactionByID(txID uuid.UUID) (*entities.Transaction, error) {
	var txn entities.Transaction
	if err := r.db.First(&txn, "id = ?", txID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(dto.MESSAGE_TRANSACTION_NOT_FOUND)
		}
		return nil, err
	}
	return &txn, nil
}

func (r *tripRepository) TransactionWithRelations(txID uuid.UUID) (*entities.Transaction, error) {
	var txn entities.Transaction
	if err := r.db.
		Preload("Customer").
		Preload("Driver").
		Preload("Vehicle").
		First(&txn, "id = ?", txID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(dto.MESSAGE_TRANSACTION_NOT_FOUND)
		}
		return nil, err
	}
	return &txn, nil
}

func (r *tripRepository) UpdateStatus(txID uuid.UUID, from, to entities.TransactionStatus, extra map[string]any) (bool, error) {
	updates := map[string]any{"status": to}
	maps.Copy(updates, extra)

	result := r.db.Model(&entities.Transaction{}).
		Where("id = ? AND status = ?", txID, from).
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *tripRepository) UpdateDriverLastLatLong(txID uuid.UUID, lat, lng string) error {
	return r.db.Model(&entities.Transaction{}).
		Where("id = ?", txID).
		Update("driver_last_lat_long", []string{lat, lng}).Error
}

func (r *tripRepository) UpdateCustomerLastLatLong(txID uuid.UUID, lat, lng string) error {
	return r.db.Model(&entities.Transaction{}).
		Where("id = ?", txID).
		Update("customer_last_lat_long", []string{lat, lng}).Error
}

func CompleteExtra(paidAt time.Time) map[string]any {
	return map[string]any{"paid_at": paidAt}
}
