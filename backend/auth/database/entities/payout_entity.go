package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "pending"
	PayoutStatusProcessing PayoutStatus = "processing"
	PayoutStatusCancelled  PayoutStatus = "cancelled"
	PayoutStatusPaid       PayoutStatus = "paid"
	PayoutStatusFailed     PayoutStatus = "failed"
)

type Payout struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`

	DriverID uuid.UUID `gorm:"type:uuid;not null" json:"driver_id"`
	Driver   Driver    `gorm:"foreignKey:DriverID;references:ID;constraint:OnDelete:CASCADE" json:"driver"`

	Amount       int          `gorm:"type:int;not null" json:"amount"`
	Status       PayoutStatus `gorm:"type:payout_status;not null" json:"status"`
	FailedReason *string      `gorm:"type:text" json:"failed_reason"`

	Timestamp
}

func (Payout) TableName() string {
	return "payouts"
}

func (p *Payout) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
