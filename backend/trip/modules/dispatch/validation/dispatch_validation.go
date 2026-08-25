package validation

import (
	"github.com/go-playground/validator/v10"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/dto"
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
	validate.RegisterValidation("driver_mode", validation.Enum(
		dto.DriverModeOnline,
		dto.DriverModeOffline,
	))
	validate.RegisterValidation("offer_action", validation.Enum(
		dto.OfferActionAccept,
		dto.OfferActionReject,
	))

	return &DispatchValidation{
		validate: validate,
	}
}

func (v *DispatchValidation) ValidateRequest(req any) error {
	return v.validate.Struct(req)
}
