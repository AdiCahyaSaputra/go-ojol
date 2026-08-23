package tests

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/service"
	casbinrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/casbin/repository"
	userDto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/repository"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/helpers"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubEnforcer struct {
	loadPolicyCalled bool
}

func (s *stubEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	return true, nil
}

func (s *stubEnforcer) LoadPolicy() error {
	s.loadPolicyCalled = true
	return nil
}

type stubUploadClient struct{}

func (s *stubUploadClient) Upload(ctx context.Context, filename, contentType string, size int64, body io.Reader) (string, error) {
	_, _ = io.Copy(io.Discard, body)
	return "https://example.com/" + filename, nil
}

func setupAuthTestDB(t *testing.T) *gorm.DB {
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
		CREATE TABLE casbin_rules (
			id TEXT PRIMARY KEY,
			ptype TEXT NOT NULL,
			v0 TEXT,
			v1 TEXT,
			v2 TEXT,
			v3 TEXT,
			v4 TEXT,
			v5 TEXT,
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

func seedVehicle(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO vehicles (id, name, license_number, max_size, type) VALUES (?, ?, ?, ?, ?)`,
		id.String(), "Test Bike", "B1234XYZ", 2, "motorcycle",
	).Error)
	return id
}

func newAuthServiceForTest(t *testing.T, db *gorm.DB, enforcer *stubEnforcer) service.AuthService {
	t.Helper()

	jwtService, _ := newTestJWTService(t)
	return service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		jwtService,
		enforcer,
		&stubUploadClient{},
		db,
	)
}

func TestAuthService_Register_CreatesGroupingByEmail(t *testing.T) {
	db := setupAuthTestDB(t)
	enforcer := &stubEnforcer{}
	svc := newAuthServiceForTest(t, db, enforcer)
	ctx := context.Background()

	result, err := svc.Register(ctx, userDto.UserCreateRequest{
		Email:    "ada@example.com",
		Password: "password123",
		Role:     constants.ENUM_ROLE_CUSTOMER,
		Name:     "Ada",
	})
	require.NoError(t, err)
	assert.Equal(t, "ada@example.com", result.Email)
	assert.Equal(t, constants.ENUM_ROLE_CUSTOMER, result.Role)
	require.NotNil(t, result.Customer)
	assert.Equal(t, "Ada", result.Customer.Name)
	assert.True(t, enforcer.loadPolicyCalled)

	var rule entities.CasbinRule
	err = db.Where("ptype = ? AND v0 = ?", entities.CasbinRulePtypeRole, "ada@example.com").First(&rule).Error
	require.NoError(t, err)
	assert.Equal(t, constants.ENUM_ROLE_CUSTOMER, rule.V1)
}

func TestAuthService_Register_DriverRole(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})

	result, err := svc.Register(context.Background(), userDto.UserCreateRequest{
		Email:                "bob@example.com",
		Password:             "password123",
		Role:                 constants.ENUM_ROLE_DRIVER,
		Name:                 "Bob",
		VehicleName:          "Honda Beat",
		VehicleLicenseNumber: "B 1001 XYZ",
		VehicleMaxSize:       1,
		VehicleType:          string(entities.VehicleTypeMotorcycle),
	})
	require.NoError(t, err)
	assert.Equal(t, constants.ENUM_ROLE_DRIVER, result.Role)
	require.NotNil(t, result.Driver)
	assert.Equal(t, "Bob", result.Driver.Name)
	assert.Nil(t, result.Driver.Vehicle)

	var vehicle entities.Vehicle
	err = db.Where("license_number = ?", "B 1001 XYZ").First(&vehicle).Error
	require.NoError(t, err)
	assert.Equal(t, "Honda Beat", vehicle.Name)
	assert.Equal(t, 1, vehicle.MaxSize)
	assert.Equal(t, entities.VehicleTypeMotorcycle, vehicle.Type)

	var driver entities.Driver
	err = db.Where("user_id = ?", result.ID).First(&driver).Error
	require.NoError(t, err)
	assert.Equal(t, vehicle.ID, driver.VehicleID)
}

func TestAuthService_Register_DriverRequiresVehicle(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})

	_, err := svc.Register(context.Background(), userDto.UserCreateRequest{
		Email:    "bob@example.com",
		Password: "password123",
		Role:     constants.ENUM_ROLE_DRIVER,
	})
	assert.ErrorIs(t, err, userDto.ErrVehicleRequired)
}

func TestAuthService_Register_DriverDuplicateLicense(t *testing.T) {
	db := setupAuthTestDB(t)
	seedVehicle(t, db)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})

	_, err := svc.Register(context.Background(), userDto.UserCreateRequest{
		Email:                "bob@example.com",
		Password:             "password123",
		Role:                 constants.ENUM_ROLE_DRIVER,
		Name:                 "Bob",
		VehicleName:          "Other Bike",
		VehicleLicenseNumber: "B1234XYZ",
		VehicleMaxSize:       2,
		VehicleType:          string(entities.VehicleTypeMotorcycle),
	})
	assert.ErrorIs(t, err, userDto.ErrLicenseNumberExists)

	var userCount int64
	require.NoError(t, db.Model(&entities.User{}).Where("email = ?", "bob@example.com").Count(&userCount).Error)
	assert.Equal(t, int64(0), userCount)
}

func TestAuthService_Register_RejectsAdminRole(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})

	_, err := svc.Register(context.Background(), userDto.UserCreateRequest{
		Email:    "root@example.com",
		Password: "password123",
		Role:     constants.ENUM_ROLE_ADMIN,
	})
	assert.ErrorIs(t, err, userDto.ErrInvalidRole)
}

func TestAuthService_Login_UsesGroupingRoleAndEmail(t *testing.T) {
	db := setupAuthTestDB(t)
	jwtService, _ := newTestJWTService(t)
	svc := service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		jwtService,
		&stubEnforcer{},
		&stubUploadClient{},
		db,
	)
	ctx := context.Background()

	hashed, err := helpers.HashPassword("password123")
	require.NoError(t, err)

	user := entities.User{Email: "ada@example.com", Password: hashed}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)
	require.NoError(t, casbinrepo.NewCasbinRepository(db).AddGroupingPolicy(ctx, db, user.Email, constants.ENUM_ROLE_DRIVER))

	result, err := svc.Login(ctx, userDto.UserLoginRequest{
		Email:    "ada@example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, constants.ENUM_ROLE_DRIVER, result.Role)

	parser := new(jwt.Parser)
	unverified, _, err := parser.ParseUnverified(result.AccessToken, jwt.MapClaims{})
	require.NoError(t, err)
	claims := unverified.Claims.(jwt.MapClaims)
	assert.Equal(t, "ada@example.com", claims["email"])
	assert.Equal(t, constants.ENUM_ROLE_DRIVER, claims["role"])
}

func TestAuthService_Login_RejectsWhenRoleMissing(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})
	ctx := context.Background()

	hashed, err := helpers.HashPassword("password123")
	require.NoError(t, err)

	user := entities.User{Email: "ghost@example.com", Password: hashed}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)

	_, err = svc.Login(ctx, userDto.UserLoginRequest{
		Email:    "ghost@example.com",
		Password: "password123",
	})
	assert.ErrorIs(t, err, userDto.ErrRoleNotAssigned)
}

func TestAuthService_Register_RejectsDuplicateEmail(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})
	ctx := context.Background()

	req := userDto.UserCreateRequest{
		Email:    "ada@example.com",
		Password: "password123",
		Role:     constants.ENUM_ROLE_CUSTOMER,
	}
	_, err := svc.Register(ctx, req)
	require.NoError(t, err)

	_, err = svc.Register(ctx, req)
	assert.ErrorIs(t, err, userDto.ErrEmailAlreadyExists)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})
	ctx := context.Background()

	hashed, err := helpers.HashPassword("password123")
	require.NoError(t, err)
	user := entities.User{Email: "ada@example.com", Password: hashed}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)

	_, err = svc.Login(ctx, userDto.UserLoginRequest{
		Email:    "ada@example.com",
		Password: "wrong-password",
	})
	assert.ErrorIs(t, err, dto.ErrInvalidCredentials)
}
