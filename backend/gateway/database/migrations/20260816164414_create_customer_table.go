package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260816164414_create_customer_table", UpCreateCustomerTable, DownCreateCustomerTable)
}

func UpCreateCustomerTable(db *gorm.DB) error {
	return db.AutoMigrate(&entities.Customer{})
}

func DownCreateCustomerTable(db *gorm.DB) error {
	return db.Migrator().DropTable(&entities.Customer{})
}
