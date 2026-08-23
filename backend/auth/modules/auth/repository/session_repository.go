package repository

import (
	"context"
	"time"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(ctx context.Context, tx *gorm.DB, session entities.Session) (entities.Session, error)
	FindByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Session, error)
	FindActiveByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Session, error)
	FindByRefreshTokenHash(ctx context.Context, tx *gorm.DB, hash string) (entities.Session, error)
	UpdateRefreshToken(ctx context.Context, tx *gorm.DB, id uuid.UUID, hash string, expiresAt time.Time) error
	ListByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) ([]entities.Session, error)
	RevokeByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) dbOr(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *sessionRepository) Create(ctx context.Context, tx *gorm.DB, session entities.Session) (entities.Session, error) {
	if err := r.dbOr(tx).WithContext(ctx).Create(&session).Error; err != nil {
		return entities.Session{}, err
	}
	return session, nil
}

func (r *sessionRepository) FindByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Session, error) {
	var session entities.Session
	err := r.dbOr(tx).WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		return entities.Session{}, err
	}
	return session, nil
}

func (r *sessionRepository) FindActiveByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Session, error) {
	var session entities.Session
	err := r.dbOr(tx).WithContext(ctx).
		Where("id = ? AND revoked_at IS NULL AND expires_at > ?", id, time.Now()).
		First(&session).Error
	if err != nil {
		return entities.Session{}, err
	}
	return session, nil
}

func (r *sessionRepository) FindByRefreshTokenHash(ctx context.Context, tx *gorm.DB, hash string) (entities.Session, error) {
	var session entities.Session
	err := r.dbOr(tx).WithContext(ctx).Where("refresh_token_hash = ?", hash).First(&session).Error
	if err != nil {
		return entities.Session{}, err
	}
	return session, nil
}

func (r *sessionRepository) UpdateRefreshToken(ctx context.Context, tx *gorm.DB, id uuid.UUID, hash string, expiresAt time.Time) error {
	return r.dbOr(tx).WithContext(ctx).Model(&entities.Session{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"refresh_token_hash": hash,
			"expires_at":         expiresAt,
		}).Error
}

func (r *sessionRepository) ListByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) ([]entities.Session, error) {
	var sessions []entities.Session
	err := r.dbOr(tx).WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *sessionRepository) RevokeByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	now := time.Now()
	return r.dbOr(tx).WithContext(ctx).Model(&entities.Session{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now).Error
}

func (r *sessionRepository) RevokeAllByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID) error {
	now := time.Now()
	return r.dbOr(tx).WithContext(ctx).Model(&entities.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}
