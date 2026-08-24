package repository

import (
	"context"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"gorm.io/gorm"
)

type (
	UserRepository interface {
		Register(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error)
		CreateCustomer(ctx context.Context, tx *gorm.DB, customer entities.Customer) (entities.Customer, error)
		CreateDriver(ctx context.Context, tx *gorm.DB, driver entities.Driver) (entities.Driver, error)
		FindVehicleByLicenseNumber(ctx context.Context, tx *gorm.DB, licenseNumber string) (*entities.Vehicle, error)
		CreateVehicle(ctx context.Context, tx *gorm.DB, vehicle entities.Vehicle) (entities.Vehicle, error)
		GetUserById(ctx context.Context, tx *gorm.DB, userId string) (entities.User, error)
		GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, error)
		CheckEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, bool, error)
		Update(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error)
		Delete(ctx context.Context, tx *gorm.DB, userId string) error
	}

	userRepository struct {
		db *gorm.DB
	}
)

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Register(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) CreateCustomer(ctx context.Context, tx *gorm.DB, customer entities.Customer) (entities.Customer, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&customer).Error; err != nil {
		return entities.Customer{}, err
	}

	return customer, nil
}

func (r *userRepository) CreateDriver(ctx context.Context, tx *gorm.DB, driver entities.Driver) (entities.Driver, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&driver).Error; err != nil {
		return entities.Driver{}, err
	}

	return driver, nil
}

func (r *userRepository) FindVehicleByLicenseNumber(ctx context.Context, tx *gorm.DB, licenseNumber string) (*entities.Vehicle, error) {
	if tx == nil {
		tx = r.db
	}

	var vehicle entities.Vehicle
	err := tx.WithContext(ctx).Where("license_number = ?", licenseNumber).Take(&vehicle).Error
	if err != nil {
		return nil, err
	}

	return &vehicle, nil
}

func (r *userRepository) CreateVehicle(ctx context.Context, tx *gorm.DB, vehicle entities.Vehicle) (entities.Vehicle, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&vehicle).Error; err != nil {
		return entities.Vehicle{}, err
	}

	return vehicle, nil
}

func (r *userRepository) GetUserById(ctx context.Context, tx *gorm.DB, userId string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).
		Preload("Customer").
		Preload("Driver.Vehicle").
		Where("id = ?", userId).
		Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) CheckEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, bool, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		return entities.User{}, false, err
	}

	return user, true, nil
}

func (r *userRepository) Update(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Updates(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) Delete(ctx context.Context, tx *gorm.DB, userId string) error {
	if tx == nil {
		tx = r.db
	}

	result := tx.WithContext(ctx).Delete(&entities.User{}, "id = ?", userId)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
