package utils

import (
	"errors"

	pkgdto "github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/dto"
)

func ClientErrorMessage(err error, known ...error) any {
	if err == nil {
		return nil
	}
	for _, candidate := range known {
		if errors.Is(err, candidate) {
			return err.Error()
		}
	}
	return pkgdto.MESSAGE_INTERNAL_SERVER_ERROR
}
