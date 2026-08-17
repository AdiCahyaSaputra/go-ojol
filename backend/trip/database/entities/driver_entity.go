package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Driver struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User   User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"user"`

	VehicleID uuid.UUID `gorm:"type:uuid;not null" json:"vehicle_id"`
	Vehicle   Vehicle   `gorm:"foreignKey:VehicleID;references:ID;constraint:OnDelete:SET NULL" json:"vehicle"`

	Name              string  `gorm:"type:varchar(255);not null" json:"name"`
	PhoneNumber       string  `gorm:"type:varchar(15);not null" json:"phone_number"`
	Address           string  `gorm:"type:text;not null" json:"address"`
	ProfilePictureUrl *string `gorm:"type:text" json:"profile_picture_url"`

	Timestamp
}

func (Driver) TableName() string {
	return "drivers"
}

func (d *Driver) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
