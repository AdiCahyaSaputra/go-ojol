package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Checker {
	return &repository{db: db}
}

func (r *repository) IsActive(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("sessions").
		Where("id = ? AND revoked_at IS NULL AND expires_at > ?", id, time.Now()).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
