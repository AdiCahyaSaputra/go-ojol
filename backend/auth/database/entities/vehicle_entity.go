package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
)

type Vehicle struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	Name          string      `gorm:"type:varchar(150);not null" json:"name"`
	LicenseNumber string      `gorm:"type:varchar(20);not null" json:"license_number"`
	MaxSize       int         `gorm:"type:int;check:max_size > 0;not null" json:"max_size"`
	Type          VehicleType `gorm:"type:vehicle_type;not null" json:"type"`

	Timestamp
}

func (Vehicle) TableName() string {
	return "vehicles"
}

func (v *Vehicle) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}
