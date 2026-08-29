package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type TransactionStatus string

const (
	TransactionStatusPending       TransactionStatus = "pending"
	TransactionStatusOffered       TransactionStatus = "offered"
	TransactionStatusAcceptedOffer TransactionStatus = "accepted_offer"
	TransactionStatusRejectedOffer TransactionStatus = "rejected_offer"
	TransactionStatusOnTheWay      TransactionStatus = "on_the_way"
	TransactionStatusCompleted     TransactionStatus = "completed"
	TransactionStatusExpired       TransactionStatus = "expired"
	TransactionStatusCancelled     TransactionStatus = "cancelled"
)

type Transaction struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	CustomerID *uuid.UUID `gorm:"type:uuid" json:"customer_id"`
	Customer   *Customer  `gorm:"foreignKey:CustomerID;references:ID;constraint:OnDelete:SET NULL" json:"customer"`
	DriverID   *uuid.UUID `gorm:"type:uuid" json:"driver_id"`
	Driver     *Driver    `gorm:"foreignKey:DriverID;references:ID;constraint:OnDelete:SET NULL" json:"driver"`
	VehicleID  *uuid.UUID `gorm:"type:uuid" json:"vehicle_id"`
	Vehicle    *Vehicle   `gorm:"foreignKey:VehicleID;references:ID;constraint:OnDelete:SET NULL" json:"vehicle"`

	PickupLatLong        pq.StringArray `gorm:"type:varchar(40)[];not null" json:"pickup_lat_long"`
	DestinationLatLong   pq.StringArray `gorm:"type:varchar(40)[];not null" json:"destination_lat_long"`
	DriverLastLatLong    pq.StringArray `gorm:"type:varchar(40)[];not null" json:"driver_last_lat_long"`
	CustomerLastLatLong  pq.StringArray `gorm:"type:varchar(40)[]" json:"customer_last_lat_long"`

	Distance           int               `gorm:"type:int;check:distance > 0;not null" json:"distance"`
	FarePerDistance    int               `gorm:"type:int;not null" json:"fare_per_distance"`
	PlatformPercentage int               `gorm:"type:int;not null" json:"platform_percentage"`
	TotalFare          int               `gorm:"type:int;not null" json:"total_fare"`
	Status             TransactionStatus `gorm:"type:transaction_status;not null" json:"status"`
	PaidAt             *time.Time        `gorm:"type:timestamptz" json:"paid_at"`

	Timestamp
}

func (t *Transaction) TableName() string {
	return "transactions"
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
