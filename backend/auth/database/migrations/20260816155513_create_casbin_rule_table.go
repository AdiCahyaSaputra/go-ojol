package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260816155513_casbin_rules", UpCasbinRules, DownCasbinRules)
}

func UpCasbinRules(db *gorm.DB) error {
	db.Exec("create type casbin_rule_ptype as enum('p', 'g');")

	return db.AutoMigrate(&entities.CasbinRule{})
}

func DownCasbinRules(db *gorm.DB) error {
	db.Exec("drop type casbin_rule_ptype;")

	return db.Migrator().DropTable(&entities.CasbinRule{})
}
