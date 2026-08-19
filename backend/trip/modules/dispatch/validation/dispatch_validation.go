package validation

import (
	"github.com/go-playground/validator/v10"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	validation "github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/validation"
)

type DispatchValidation struct {
	validate *validator.Validate
}

func NewDispatchValidation() *DispatchValidation {
	validate := validator.New()
	validate.RegisterValidation("latlong", validation.ValidateLatLong)
	validate.RegisterValidation("vehicle_type", validation.Enum(
		entities.VehicleTypeCar,
		entities.VehicleTypeMotorcycle,
	))

	return &DispatchValidation{
		validate: validate,
	}
}

func (v *DispatchValidation) ValidateRequest(req any) error {
	return v.validate.Struct(req)
}
