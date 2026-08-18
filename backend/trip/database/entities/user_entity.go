package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Email    string    `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	Password string    `gorm:"type:varchar(255);not null" json:"password"`

	Timestamp
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(_ *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	return nil
}
