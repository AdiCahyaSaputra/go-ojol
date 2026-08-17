package tests

import (
	"context"
	"fmt"
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

	return db
}

func newAuthServiceForTest(t *testing.T, db *gorm.DB, enforcer *stubEnforcer) service.AuthService {
	t.Helper()

	jwtService, _ := newTestJWTService(t)
	return service.NewAuthService(
		repository.NewUserRepository(db),
		casbinrepo.NewCasbinRepository(db),
		jwtService,
		enforcer,
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
	})
	require.NoError(t, err)
	assert.Equal(t, "ada@example.com", result.Email)
	assert.Equal(t, constants.ENUM_ROLE_CUSTOMER, result.Role)
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
		Email:    "bob@example.com",
		Password: "password123",
		Role:     constants.ENUM_ROLE_DRIVER,
	})
	require.NoError(t, err)
	assert.Equal(t, constants.ENUM_ROLE_DRIVER, result.Role)
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
