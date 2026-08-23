package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User             User       `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	RefreshTokenHash string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	ExpiresAt        time.Time  `gorm:"type:timestamp with time zone;not null" json:"expires_at"`
	RevokedAt        *time.Time `gorm:"type:timestamp with time zone" json:"revoked_at,omitempty"`
	UserAgent        *string    `gorm:"type:varchar(512)" json:"user_agent,omitempty"`
	IP               *string    `gorm:"type:varchar(64)" json:"ip,omitempty"`

	Timestamp
}

func (Session) TableName() string {
	return "sessions"
}

func (s *Session) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *Session) IsActive(now time.Time) bool {
	return s.RevokedAt == nil && s.ExpiresAt.After(now)
}
