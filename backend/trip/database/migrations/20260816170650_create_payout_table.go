package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260816170650_create_payout_table", UpCreatePayoutTable, DownCreatePayoutTable)
}

func UpCreatePayoutTable(db *gorm.DB) error {
	db.Exec("create type payout_status as enum ('pending', 'processing', 'cancelled', 'paid', 'failed');")

	return db.AutoMigrate(&entities.Payout{})
}

func DownCreatePayoutTable(db *gorm.DB) error {
	return db.Migrator().DropTable(&entities.Payout{})
}
