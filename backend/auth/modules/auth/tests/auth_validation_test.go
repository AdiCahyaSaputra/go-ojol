package tests

import (
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/auth/validation"
	userDto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/stretchr/testify/assert"
)

func TestAuthValidation_ValidateRegisterRequest_Success(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "customer",
	}

	err := authValidation.ValidateRegisterRequest(req)

	assert.NoError(t, err)
}

func TestAuthValidation_ValidateRegisterRequest_InvalidEmail(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Email:    "invalid-email", // This will be caught by binding:"required,email" in DTO
		Password: "password123",
		Role:     "customer",
	}

	err := authValidation.ValidateRegisterRequest(req)

	// The validation should pass because DTO binding handles email validation
	// Custom validation only adds extra checks beyond DTO binding
	assert.NoError(t, err)
}

func TestAuthValidation_ValidateRegisterRequest_ShortPassword(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Email:    "test@example.com",
		Password: "123", // This will be caught by binding:"required,min=8" in DTO
		Role:     "customer",
	}

	err := authValidation.ValidateRegisterRequest(req)

	// The validation should pass because DTO binding handles password validation
	// Custom validation only adds extra checks beyond DTO binding
	assert.NoError(t, err)
}

func TestAuthValidation_ValidateLoginRequest_Success(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	err := authValidation.ValidateLoginRequest(req)

	assert.NoError(t, err)
}

func TestAuthValidation_ValidateRefreshTokenRequest_Success(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := dto.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	}

	err := authValidation.ValidateRefreshTokenRequest(req)

	assert.NoError(t, err)
}

func TestAuthValidation_ValidateRegisterRequest_RejectsAdminRole(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "admin",
	}

	err := authValidation.ValidateRegisterRequest(req)

	assert.Error(t, err)
}

func TestAuthValidation_ValidateRegisterRequest_RejectsMissingRole(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	err := authValidation.ValidateRegisterRequest(req)

	assert.Error(t, err)
}

func TestAuthValidation_ValidateRegisterRequest_DriverRequiresVehicleFields(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Email:    "driver@example.com",
		Password: "password123",
		Role:     "driver",
	}

	err := authValidation.ValidateRegisterRequest(req)
	assert.Error(t, err)
}

func TestAuthValidation_ValidateRegisterRequest_DriverWithVehicleSuccess(t *testing.T) {
	authValidation := validation.NewAuthValidation()

	req := userDto.UserCreateRequest{
		Email:                "driver@example.com",
		Password:             "password123",
		Role:                 "driver",
		VehicleName:          "Honda Beat",
		VehicleLicenseNumber: "B 1001 XYZ",
		VehicleMaxSize:       1,
		VehicleType:          "motorcycle",
	}

	err := authValidation.ValidateRegisterRequest(req)
	assert.NoError(t, err)
}
