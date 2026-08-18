package helpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)


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
			default:
				errorsMap[field] = fmt.Sprintf("%s failed validation: %s", fieldErr.Field(), tag)
			}
		}

		return errorsMap
	default:
		errorsMap["_"] = e.(string)

		return errorsMap
	}
}
