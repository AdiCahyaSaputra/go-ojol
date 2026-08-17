package tests

import (
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/validation"
	"github.com/stretchr/testify/assert"
)

func TestUserValidation_ValidateUserCreateRequest_Success(t *testing.T) {
	userValidation := validation.NewUserValidation()

	req := dto.UserCreateRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "customer",
	}

	err := userValidation.ValidateUserCreateRequest(req)

	assert.NoError(t, err)
}

func TestUserValidation_ValidateUserUpdateRequest_Success(t *testing.T) {
	userValidation := validation.NewUserValidation()

	req := dto.UserUpdateRequest{
		Email: "updated@example.com",
	}

	err := userValidation.ValidateUserUpdateRequest(req)

	assert.NoError(t, err)
}
