package helpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/go-playground/validator/v10"
)

func joinEnumStrings(enumStrings ...string) string {
	return strings.Join(enumStrings, ", ")
}

func ParseValidationError(err any) map[string]string {
	errorsMap := make(map[string]string)

	switch e := err.(type) {
	case error:

		var validationErrors validator.ValidationErrors

		if !errors.As(e, &validationErrors) {
			return errorsMap
		}

		for _, fieldErr := range validationErrors {
			field := fieldErr.Field()
			tag := fieldErr.Tag()

			field = strings.ToLower(field)

			switch tag {
			case "required":
				errorsMap[field] = fmt.Sprintf("%s is required", fieldErr.Field())
			case "email":
				errorsMap[field] = fmt.Sprintf("%s must be a valid email", fieldErr.Field())
			case "min":
				errorsMap[field] = fmt.Sprintf("%s must be at least %s characters", fieldErr.Field(), fieldErr.Param())
			case "max":
				errorsMap[field] = fmt.Sprintf("%s must be at most %s characters", fieldErr.Field(), fieldErr.Param())
			case "latlong":
				errorsMap[field] = fmt.Sprintf("%s must a valid [latitude, longitude] value", fieldErr.Field())
			case "vehicle_type":
				errorsMap[field] = fmt.Sprintf(
					"%s should be one of %s",
					fieldErr.Field(),
					joinEnumStrings(
						string(entities.VehicleTypeCar),
						string(entities.VehicleTypeMotorcycle),
					),
				)
			default:
				errorsMap[field] = fmt.Sprintf("%s failed validation: %s", fieldErr.Field(), tag)
			}
		}

		return errorsMap
	default:
		if e == nil {
			return errorsMap
		}

		errorsMap["_"] = e.(string)

		return errorsMap
	}
}
