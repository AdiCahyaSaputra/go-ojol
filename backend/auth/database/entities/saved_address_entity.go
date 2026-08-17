package entities

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type SavedAddress struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	CustomerID uuid.UUID `gorm:"type:uuid;not null" json:"customer_id"`
	Customer   Customer  `gorm:"foreignKey:CustomerID;references:ID;constraint:OnDelete:CASCADE" json:"customer"`

	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	LatLong         pq.StringArray `gorm:"type:varchar(40)[];not null" json:"lat_long"`
	IsDefaultPickup bool           `gorm:"type:boolean;not null;default:false" json:"is_default_pickup"`

	Timestamp
}

func (s *SavedAddress) TableName() string {
	return "saved_addresses"
}

func (s *SavedAddress) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
