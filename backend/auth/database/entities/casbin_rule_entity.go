package entities

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CasbinRulePtype string

const (
	CasbinRulePtypePerm CasbinRulePtype = "p"
	CasbinRulePtypeRole CasbinRulePtype = "g"
)

type CasbinRule struct {
	ID    uuid.UUID       `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Ptype CasbinRulePtype `gorm:"type:casbin_rule_ptype;not null" json:"ptype"`
	V0    string          `gorm:"type:varchar(255)" json:"v0"`
	V1    string          `gorm:"type:varchar(255)" json:"v1"`
	V2    string          `gorm:"type:varchar(255)" json:"v2"`
	V3    string          `gorm:"type:varchar(255)" json:"v3"`
	V4    string          `gorm:"type:varchar(255)" json:"v4"`
	V5    string          `gorm:"type:varchar(255)" json:"v5"`

	Timestamp
}

func (CasbinRule) TableName() string {
	return "casbin_rules"
}

func (c *CasbinRule) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
