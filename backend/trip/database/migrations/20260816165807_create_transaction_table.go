package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260816165807_create_transaction_table", UpCreateTransactionTable, DownCreateTransactionTable)
}

func UpCreateTransactionTable(db *gorm.DB) error {
	db.Exec("create type transaction_status as enum ('pending', 'on_the_way', 'completed', 'cancelled');")

	return db.AutoMigrate(&entities.Transaction{})
}

func DownCreateTransactionTable(db *gorm.DB) error {
	return db.Migrator().DropTable(&entities.Transaction{})
}
