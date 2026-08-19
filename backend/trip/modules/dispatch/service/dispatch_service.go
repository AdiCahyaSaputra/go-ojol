package service

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/modules/dispatch/repository"
	"gorm.io/gorm"
)

type DispatchService interface {
}

type dispatchService struct {
	dispatchRepository repository.DispatchRepository
	db                            *gorm.DB
}

func NewDispatchService(
	dispatchRepo repository.DispatchRepository,
	db *gorm.DB,
) DispatchService {
	return &dispatchService{
		dispatchRepository: dispatchRepo,
		db:                            db,
	}
}
