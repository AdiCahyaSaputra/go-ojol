package session

import (
	"context"

	"github.com/google/uuid"
)

type Checker interface {
	IsActive(ctx context.Context, id uuid.UUID) (bool, error)
}

type alwaysActive struct{}

func AlwaysActive() Checker {
	return alwaysActive{}
}

func (alwaysActive) IsActive(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}

type inactiveChecker struct{}

func Inactive() Checker {
	return inactiveChecker{}
}

func (inactiveChecker) IsActive(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
