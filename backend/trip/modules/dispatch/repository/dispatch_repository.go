package repository

import (
	"gorm.io/gorm"
)

type DispatchRepository interface {
}

type dispatchRepository struct {
	db *gorm.DB
}

func NewDispatchRepository(db *gorm.DB) DispatchRepository {
	return &dispatchRepository{
		db: db,
	}
}
