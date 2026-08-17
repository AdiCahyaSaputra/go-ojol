package validation

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/modules/user/dto"
	"github.com/go-playground/validator/v10"
)

type UserValidation struct {
	validate *validator.Validate
}

func NewUserValidation() *UserValidation {
	validate := validator.New()

	return &UserValidation{
		validate: validate,
	}
}

func (v *UserValidation) ValidateUserCreateRequest(req dto.UserCreateRequest) error {
	return v.validate.Struct(req)
}

func (v *UserValidation) ValidateUserUpdateRequest(req dto.UserUpdateRequest) error {
	return v.validate.Struct(req)
}
