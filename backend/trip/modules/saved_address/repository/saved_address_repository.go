package repository

import (
	"context"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SavedAddressRepository interface {
	ListByCustomer(ctx context.Context, tx *gorm.DB, customerID uuid.UUID) ([]entities.SavedAddress, error)
	GetByIDAndCustomer(ctx context.Context, tx *gorm.DB, id, customerID uuid.UUID) (entities.SavedAddress, error)
	Create(ctx context.Context, tx *gorm.DB, address entities.SavedAddress) (entities.SavedAddress, error)
	Update(ctx context.Context, tx *gorm.DB, address entities.SavedAddress) (entities.SavedAddress, error)
	Delete(ctx context.Context, tx *gorm.DB, id, customerID uuid.UUID) error
	ClearDefaultPickup(ctx context.Context, tx *gorm.DB, customerID uuid.UUID, exceptID uuid.UUID) error
}

type savedAddressRepository struct {
	db *gorm.DB
}

func NewSavedAddressRepository(db *gorm.DB) SavedAddressRepository {
	return &savedAddressRepository{db: db}
}

func (r *savedAddressRepository) dbOrTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *savedAddressRepository) ListByCustomer(ctx context.Context, tx *gorm.DB, customerID uuid.UUID) ([]entities.SavedAddress, error) {
	var addresses []entities.SavedAddress
	err := r.dbOrTx(tx).WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("is_default_pickup DESC, created_at DESC").
		Find(&addresses).Error
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *savedAddressRepository) GetByIDAndCustomer(ctx context.Context, tx *gorm.DB, id, customerID uuid.UUID) (entities.SavedAddress, error) {
	var address entities.SavedAddress
	err := r.dbOrTx(tx).WithContext(ctx).
		Where("id = ? AND customer_id = ?", id, customerID).
		First(&address).Error
	if err != nil {
		return entities.SavedAddress{}, err
	}
	return address, nil
}

func (r *savedAddressRepository) Create(ctx context.Context, tx *gorm.DB, address entities.SavedAddress) (entities.SavedAddress, error) {
	if err := r.dbOrTx(tx).WithContext(ctx).Create(&address).Error; err != nil {
		return entities.SavedAddress{}, err
	}
	return address, nil
}

func (r *savedAddressRepository) Update(ctx context.Context, tx *gorm.DB, address entities.SavedAddress) (entities.SavedAddress, error) {
	if err := r.dbOrTx(tx).WithContext(ctx).Save(&address).Error; err != nil {
		return entities.SavedAddress{}, err
	}
	return address, nil
}

func (r *savedAddressRepository) Delete(ctx context.Context, tx *gorm.DB, id, customerID uuid.UUID) error {
	result := r.dbOrTx(tx).WithContext(ctx).
		Where("id = ? AND customer_id = ?", id, customerID).
		Delete(&entities.SavedAddress{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *savedAddressRepository) ClearDefaultPickup(ctx context.Context, tx *gorm.DB, customerID uuid.UUID, exceptID uuid.UUID) error {
	q := r.dbOrTx(tx).WithContext(ctx).
		Model(&entities.SavedAddress{}).
		Where("customer_id = ? AND is_default_pickup = ?", customerID, true)
	if exceptID != uuid.Nil {
		q = q.Where("id <> ?", exceptID)
	}
	return q.Update("is_default_pickup", false).Error
}
