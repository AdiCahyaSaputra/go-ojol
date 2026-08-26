package validation

import (
	"github.com/go-playground/validator/v10"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/saved_address/dto"
	pkgvalidation "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/validation"
)

type SavedAddressValidation struct {
	validate *validator.Validate
}

func NewSavedAddressValidation() *SavedAddressValidation {
	validate := validator.New()
	validate.RegisterValidation("latlong", pkgvalidation.ValidateLatLong)

	return &SavedAddressValidation{
		validate: validate,
	}
}

func (v *SavedAddressValidation) ValidateCreateRequest(req dto.SavedAddressCreateRequest) error {
	return v.validate.Struct(req)
}

func (v *SavedAddressValidation) ValidateUpdateRequest(req dto.SavedAddressUpdateRequest) error {
	return v.validate.Struct(req)
}
