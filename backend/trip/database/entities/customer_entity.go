package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Customer struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"user"`

	Name              string  `gorm:"type:varchar(255);not null" json:"name"`
	PhoneNumber       string  `gorm:"type:varchar(15);not null" json:"phone_number"`
	ProfilePictureUrl *string `gorm:"type:text" json:"profile_picture_url"`

	Timestamp
}

func (c *Customer) TableName() string {
	return "customers"
}

func (c *Customer) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
