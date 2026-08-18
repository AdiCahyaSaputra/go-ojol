package utils

import "github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/helpers"

type Response struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Error   any    `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type EmptyObj struct{}

func BuildResponseSuccess(message string, data any) Response {
	res := Response{
		Status:  true,
		Message: message,
		Data:    data,
	}
	return res
}

func BuildResponseFailed(message string, err any, data any) Response {
	parsedErrors := helpers.ParseValidationError(err)

	res := Response{
		Status:  false,
		Message: message,
		Data:    data,
	}

	if errors := parsedErrors["_"]; errors != "" {
		res.Error = errors
	} else {
		res.Error = parsedErrors
	}

	return res
}
