package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/repository"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE vehicles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			license_number TEXT NOT NULL,
			max_size INTEGER NOT NULL,
			type TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE customers (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			phone_number TEXT NOT NULL,
			profile_picture_url TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE drivers (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			vehicle_id TEXT NOT NULL,
			name TEXT NOT NULL,
			phone_number TEXT NOT NULL,
			address TEXT NOT NULL,
			profile_picture_url TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)

	return db
}

func TestUserService_GetUserById_IncludesOwnedVehicle(t *testing.T) {
	db := setupUserTestDB(t)
	svc := service.NewUserService(repository.NewUserRepository(db), db)

	userID := uuid.New()
	vehicleID := uuid.New()
	driverID := uuid.New()

	require.NoError(t, db.Exec(
		`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`,
		userID.String(), "drv@example.com", "hashed",
	).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO vehicles (id, name, license_number, max_size, type) VALUES (?, ?, ?, ?, ?)`,
		vehicleID.String(), "Honda Beat", "B 1001 XYZ", 1, "motorcycle",
	).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO drivers (id, user_id, vehicle_id, name, phone_number, address) VALUES (?, ?, ?, ?, ?, ?)`,
		driverID.String(), userID.String(), vehicleID.String(), "Bob", "08123456789", "Jakarta",
	).Error)

	result, err := svc.GetUserById(context.Background(), userID.String())
	require.NoError(t, err)
	require.NotNil(t, result.Driver)
	require.NotNil(t, result.Driver.Vehicle)
	assert.Equal(t, vehicleID.String(), result.Driver.Vehicle.ID)
	assert.Equal(t, "Honda Beat", result.Driver.Vehicle.Name)
	assert.Equal(t, "B 1001 XYZ", result.Driver.Vehicle.LicenseNumber)
	assert.Equal(t, 1, result.Driver.Vehicle.MaxSize)
	assert.Equal(t, "motorcycle", result.Driver.Vehicle.Type)
}

func TestUserService_GetUserById_CustomerHasNoVehicle(t *testing.T) {
	db := setupUserTestDB(t)
	svc := service.NewUserService(repository.NewUserRepository(db), db)

	userID := uuid.New()
	customerID := uuid.New()

	require.NoError(t, db.Exec(
		`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`,
		userID.String(), "cst@example.com", "hashed",
	).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO customers (id, user_id, name, phone_number) VALUES (?, ?, ?, ?)`,
		customerID.String(), userID.String(), "Ada", "08111111111",
	).Error)

	result, err := svc.GetUserById(context.Background(), userID.String())
	require.NoError(t, err)
	require.NotNil(t, result.Customer)
	assert.Nil(t, result.Driver)
}

func TestUserService_DeleteRemovesTargetUser(t *testing.T) {
	db := setupUserTestDB(t)
	svc := service.NewUserService(repository.NewUserRepository(db), db)

	adminID := uuid.New()
	targetID := uuid.New()

	require.NoError(t, db.Exec(
		`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`,
		adminID.String(), "admin@example.com", "hashed",
	).Error)
	require.NoError(t, db.Exec(
		`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`,
		targetID.String(), "target@example.com", "hashed",
	).Error)

	err := svc.Delete(context.Background(), targetID.String())
	require.NoError(t, err)

	_, err = svc.GetUserById(context.Background(), targetID.String())
	assert.Error(t, err)

	_, err = svc.GetUserById(context.Background(), adminID.String())
	require.NoError(t, err)
}
