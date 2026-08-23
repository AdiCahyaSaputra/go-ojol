package tests

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/dto"
	authrepo "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/repository"
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
	require.NoError(t, db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			refresh_token_hash TEXT NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			revoked_at DATETIME,
			user_agent TEXT,
			ip TEXT,
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

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newAuthServiceForTest(t *testing.T, db *gorm.DB, enforcer *stubEnforcer) service.AuthService {
	t.Helper()

	jwtService, _ := newTestJWTService(t)
	return service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		authrepo.NewSessionRepository(db),
		jwtService,
		enforcer,
		&stubUploadClient{},
		db,
	)
}

func seedLoginUser(t *testing.T, db *gorm.DB, email, role string) entities.User {
	t.Helper()
	hashed, err := helpers.HashPassword("password123")
	require.NoError(t, err)

	user := entities.User{ID: uuid.New(), Email: email, Password: hashed}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)
	require.NoError(t, casbinrepo.NewCasbinRepository(db).AddGroupingPolicy(context.Background(), db, user.Email, role))
	return user
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

	var rule entities.CasbinRule
	err = db.Where("ptype = ? AND v0 = ?", entities.CasbinRulePtypeRole, "ada@example.com").First(&rule).Error
	require.NoError(t, err)
	assert.Equal(t, constants.ENUM_ROLE_CUSTOMER, rule.V1)
	assert.True(t, enforcer.loadPolicyCalled)
}

func TestAuthService_Register_CreatesDriverAndVehicle(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})
	ctx := context.Background()

	result, err := svc.Register(ctx, userDto.UserCreateRequest{
		Email:                "driver@example.com",
		Password:             "password123",
		Role:                 constants.ENUM_ROLE_DRIVER,
		Name:                 "Bob",
		PhoneNumber:          "081234",
		VehicleName:          "Beat",
		VehicleLicenseNumber: "B9999XYZ",
		VehicleMaxSize:       2,
		VehicleType:          string(entities.VehicleTypeMotorcycle),
	})
	require.NoError(t, err)
	assert.Equal(t, constants.ENUM_ROLE_DRIVER, result.Role)
	require.NotNil(t, result.Driver)

	var vehicle entities.Vehicle
	require.NoError(t, db.Where("license_number = ?", "B9999XYZ").First(&vehicle).Error)
	assert.Equal(t, entities.VehicleTypeMotorcycle, vehicle.Type)

	var driver entities.Driver
	err = db.Where("user_id = ?", result.ID).First(&driver).Error
	require.NoError(t, err)
	assert.Equal(t, vehicle.ID, driver.VehicleID)
}

func TestAuthService_Register_RejectsDuplicateLicense(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})
	ctx := context.Background()

	seedVehicle(t, db)

	_, err := svc.Register(ctx, userDto.UserCreateRequest{
		Email:                "bob@example.com",
		Password:             "password123",
		Role:                 constants.ENUM_ROLE_DRIVER,
		Name:                 "Bob",
		VehicleName:          "Beat",
		VehicleLicenseNumber: "B1234XYZ",
		VehicleMaxSize:       2,
		VehicleType:          string(entities.VehicleTypeMotorcycle),
	})
	assert.ErrorIs(t, err, userDto.ErrLicenseNumberExists)

	var userCount int64
	require.NoError(t, db.Model(&entities.User{}).Where("email = ?", "bob@example.com").Count(&userCount).Error)
	assert.Equal(t, int64(0), userCount)
}

func TestAuthService_Login_UsesGroupingRoleAndEmail(t *testing.T) {
	db := setupAuthTestDB(t)
	jwtService, _ := newTestJWTService(t)
	svc := service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		authrepo.NewSessionRepository(db),
		jwtService,
		&stubEnforcer{},
		&stubUploadClient{},
		db,
	)
	ctx := context.Background()

	_ = seedLoginUser(t, db, "ada@example.com", constants.ENUM_ROLE_DRIVER)

	result, err := svc.Login(ctx, userDto.UserLoginRequest{
		Email:    "ada@example.com",
		Password: "password123",
	}, dto.LoginMeta{UserAgent: "test-agent", IP: "127.0.0.1"})
	require.NoError(t, err)
	assert.Equal(t, constants.ENUM_ROLE_DRIVER, result.Role)
	assert.NotEmpty(t, result.RefreshToken)

	parser := new(jwt.Parser)
	unverified, _, err := parser.ParseUnverified(result.AccessToken, jwt.MapClaims{})
	require.NoError(t, err)
	claims := unverified.Claims.(jwt.MapClaims)
	assert.Equal(t, "ada@example.com", claims["email"])
	assert.Equal(t, constants.ENUM_ROLE_DRIVER, claims["role"])
	assert.NotEmpty(t, claims["session_id"])

	var session entities.Session
	require.NoError(t, db.Where("refresh_token_hash = ?", hashToken(result.RefreshToken)).First(&session).Error)
	assert.Equal(t, claims["session_id"], session.ID.String())
	require.NotNil(t, session.UserAgent)
	assert.Equal(t, "test-agent", *session.UserAgent)
}

func TestAuthService_Login_RejectsWhenRoleMissing(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})
	ctx := context.Background()

	hashed, err := helpers.HashPassword("password123")
	require.NoError(t, err)

	user := entities.User{ID: uuid.New(), Email: "ghost@example.com", Password: hashed}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)

	_, err = svc.Login(ctx, userDto.UserLoginRequest{
		Email:    "ghost@example.com",
		Password: "password123",
	}, dto.LoginMeta{})
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
	user := entities.User{ID: uuid.New(), Email: "ada@example.com", Password: hashed}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)

	_, err = svc.Login(ctx, userDto.UserLoginRequest{
		Email:    "ada@example.com",
		Password: "wrong-password",
	}, dto.LoginMeta{})
	assert.ErrorIs(t, err, dto.ErrInvalidCredentials)
}

func TestAuthService_Login_CreatesSeparateSessionsPerDevice(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})
	ctx := context.Background()
	user := seedLoginUser(t, db, "multi@example.com", constants.ENUM_ROLE_CUSTOMER)

	first, err := svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{UserAgent: "phone"})
	require.NoError(t, err)
	second, err := svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{UserAgent: "laptop"})
	require.NoError(t, err)

	assert.NotEqual(t, first.RefreshToken, second.RefreshToken)

	var count int64
	require.NoError(t, db.Model(&entities.Session{}).Where("user_id = ?", user.ID).Count(&count).Error)
	assert.Equal(t, int64(2), count)
}

func TestAuthService_Refresh_RotatesTokenKeepsSession(t *testing.T) {
	db := setupAuthTestDB(t)
	jwtService, _ := newTestJWTService(t)
	svc := service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		authrepo.NewSessionRepository(db),
		jwtService,
		&stubEnforcer{},
		&stubUploadClient{},
		db,
	)
	ctx := context.Background()
	_ = seedLoginUser(t, db, "refresh@example.com", constants.ENUM_ROLE_CUSTOMER)

	login, err := svc.Login(ctx, userDto.UserLoginRequest{Email: "refresh@example.com", Password: "password123"}, dto.LoginMeta{})
	require.NoError(t, err)

	oldSessionID, err := jwtService.GetSessionIDByToken(login.AccessToken)
	require.NoError(t, err)

	refreshed, err := svc.Refresh(ctx, dto.RefreshTokenRequest{RefreshToken: login.RefreshToken})
	require.NoError(t, err)
	assert.NotEqual(t, login.RefreshToken, refreshed.RefreshToken)
	assert.NotEqual(t, login.AccessToken, refreshed.AccessToken)

	newSessionID, err := jwtService.GetSessionIDByToken(refreshed.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, oldSessionID, newSessionID)

	_, err = svc.Refresh(ctx, dto.RefreshTokenRequest{RefreshToken: login.RefreshToken})
	assert.ErrorIs(t, err, dto.ErrRefreshTokenNotFound)
}

func TestAuthService_RevokeSession_InvalidatesAccessCheck(t *testing.T) {
	db := setupAuthTestDB(t)
	jwtService, _ := newTestJWTService(t)
	sessionRepo := authrepo.NewSessionRepository(db)
	svc := service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		sessionRepo,
		jwtService,
		&stubEnforcer{},
		&stubUploadClient{},
		db,
	)
	ctx := context.Background()
	user := seedLoginUser(t, db, "revoke@example.com", constants.ENUM_ROLE_CUSTOMER)

	deviceA, err := svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{UserAgent: "A"})
	require.NoError(t, err)
	deviceB, err := svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{UserAgent: "B"})
	require.NoError(t, err)

	sessionAID, err := jwtService.GetSessionIDByToken(deviceA.AccessToken)
	require.NoError(t, err)
	sessionBID, err := jwtService.GetSessionIDByToken(deviceB.AccessToken)
	require.NoError(t, err)

	require.NoError(t, svc.RevokeSession(ctx, user.ID.String(), sessionAID))

	sidA, err := uuid.Parse(sessionAID)
	require.NoError(t, err)
	_, err = sessionRepo.FindActiveByID(ctx, db, sidA)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	sidB, err := uuid.Parse(sessionBID)
	require.NoError(t, err)
	_, err = sessionRepo.FindActiveByID(ctx, db, sidB)
	require.NoError(t, err)

	_, err = svc.Refresh(ctx, dto.RefreshTokenRequest{RefreshToken: deviceA.RefreshToken})
	assert.ErrorIs(t, err, dto.ErrSessionRevoked)
}

func TestAuthService_Logout_RevokesCurrentSession(t *testing.T) {
	db := setupAuthTestDB(t)
	jwtService, _ := newTestJWTService(t)
	sessionRepo := authrepo.NewSessionRepository(db)
	svc := service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		sessionRepo,
		jwtService,
		&stubEnforcer{},
		&stubUploadClient{},
		db,
	)
	ctx := context.Background()
	_ = seedLoginUser(t, db, "logout@example.com", constants.ENUM_ROLE_CUSTOMER)

	tokens, err := svc.Login(ctx, userDto.UserLoginRequest{Email: "logout@example.com", Password: "password123"}, dto.LoginMeta{})
	require.NoError(t, err)
	sessionID, err := jwtService.GetSessionIDByToken(tokens.AccessToken)
	require.NoError(t, err)

	require.NoError(t, svc.Logout(ctx, sessionID))

	sid, err := uuid.Parse(sessionID)
	require.NoError(t, err)
	_, err = sessionRepo.FindActiveByID(ctx, db, sid)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestAuthService_LogoutAll_RevokesEverySession(t *testing.T) {
	db := setupAuthTestDB(t)
	svc := newAuthServiceForTest(t, db, &stubEnforcer{})
	ctx := context.Background()
	user := seedLoginUser(t, db, "all@example.com", constants.ENUM_ROLE_CUSTOMER)

	_, err := svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{})
	require.NoError(t, err)
	_, err = svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{})
	require.NoError(t, err)

	require.NoError(t, svc.LogoutAll(ctx, user.ID.String()))

	var active int64
	require.NoError(t, db.Model(&entities.Session{}).Where("user_id = ? AND revoked_at IS NULL", user.ID).Count(&active).Error)
	assert.Equal(t, int64(0), active)
}

func TestAuthService_ListSessions_MarksCurrent(t *testing.T) {
	db := setupAuthTestDB(t)
	jwtService, _ := newTestJWTService(t)
	svc := service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		authrepo.NewSessionRepository(db),
		jwtService,
		&stubEnforcer{},
		&stubUploadClient{},
		db,
	)
	ctx := context.Background()
	user := seedLoginUser(t, db, "list@example.com", constants.ENUM_ROLE_CUSTOMER)

	first, err := svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{})
	require.NoError(t, err)
	_, err = svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{})
	require.NoError(t, err)

	currentID, err := jwtService.GetSessionIDByToken(first.AccessToken)
	require.NoError(t, err)

	sessions, err := svc.ListSessions(ctx, user.ID.String(), currentID)
	require.NoError(t, err)
	require.Len(t, sessions, 2)

	var currentCount int
	for _, session := range sessions {
		if session.IsCurrent {
			currentCount++
			assert.Equal(t, currentID, session.ID)
		}
	}
	assert.Equal(t, 1, currentCount)
}

func TestAuthService_Refresh_ExpiredToken(t *testing.T) {
	db := setupAuthTestDB(t)
	jwtService, _ := newTestJWTService(t)
	sessionRepo := authrepo.NewSessionRepository(db)
	svc := service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		sessionRepo,
		jwtService,
		&stubEnforcer{},
		&stubUploadClient{},
		db,
	)
	ctx := context.Background()
	user := seedLoginUser(t, db, "expired@example.com", constants.ENUM_ROLE_CUSTOMER)

	tokens, err := svc.Login(ctx, userDto.UserLoginRequest{Email: user.Email, Password: "password123"}, dto.LoginMeta{})
	require.NoError(t, err)

	require.NoError(t, db.Model(&entities.Session{}).
		Where("refresh_token_hash = ?", hashToken(tokens.RefreshToken)).
		Update("expires_at", time.Now().Add(-time.Minute)).Error)

	_, err = svc.Refresh(ctx, dto.RefreshTokenRequest{RefreshToken: tokens.RefreshToken})
	assert.ErrorIs(t, err, dto.ErrRefreshTokenExpired)
}
