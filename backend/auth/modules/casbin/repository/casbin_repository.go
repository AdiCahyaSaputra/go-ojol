package repository

import (
	"context"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"gorm.io/gorm"
)

type CasbinRepository interface {
	AddGroupingPolicy(ctx context.Context, tx *gorm.DB, email, role string) error
	GetRolesForUser(ctx context.Context, tx *gorm.DB, email string) ([]string, error)
}

type casbinRepository struct {
	db *gorm.DB
}

func NewCasbinRepository(db *gorm.DB) CasbinRepository {
	return &casbinRepository{db: db}
}

func (r *casbinRepository) AddGroupingPolicy(ctx context.Context, tx *gorm.DB, email, role string) error {
	if tx == nil {
		tx = r.db
	}

	rule := entities.CasbinRule{
		Ptype: entities.CasbinRulePtypeRole,
		V0:    email,
		V1:    role,
	}

	return tx.WithContext(ctx).Create(&rule).Error
}

func (r *casbinRepository) GetRolesForUser(ctx context.Context, tx *gorm.DB, email string) ([]string, error) {
	if tx == nil {
		tx = r.db
	}

	var rules []entities.CasbinRule
	if err := tx.WithContext(ctx).
		Where("ptype = ? AND v0 = ?", entities.CasbinRulePtypeRole, email).
		Find(&rules).Error; err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(rules))
	for _, rule := range rules {
		if rule.V1 != "" {
			roles = append(roles, rule.V1)
		}
	}

	return roles, nil
}
