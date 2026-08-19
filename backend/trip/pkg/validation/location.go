package validation

import (
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

func ValidateLatLong(fl validator.FieldLevel) bool {
	field := fl.Field()

	isNotValidKind := field.Kind() != reflect.Array && field.Kind() != reflect.Slice 
	isNotEnoughItem := field.Len() != 2

	if isNotValidKind || isNotEnoughItem {
		return false
	}

	lat, errLat := strconv.ParseFloat(field.Index(0).String(), 64)
	long, errLong := strconv.ParseFloat(field.Index(1).String(), 64)

	if errLat != nil || errLong != nil {
		return false 
	}

	return lat >= -90 && lat <= 90 && long >= -180 && long <= 180
}
