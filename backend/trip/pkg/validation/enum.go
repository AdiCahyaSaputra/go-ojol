package validation

import "github.com/go-playground/validator/v10"

func Enum[T ~string](allowed ...T) validator.Func {
	set := make(map[string]struct{}, len(allowed))

	for _, v := range allowed {
		set[string(v)] = struct{}{}
	}

	return func(fl validator.FieldLevel) bool {
		_, ok := set[fl.Field().String()]

		return ok
	}
}
