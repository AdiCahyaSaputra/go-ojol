package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/gateway/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260816165252_create_saved_address_table", UpCreateSavedAddressTable, DownCreateSavedAddressTable)
}

func UpCreateSavedAddressTable(db *gorm.DB) error {
	return db.AutoMigrate(&entities.SavedAddress{})
}

func DownCreateSavedAddressTable(db *gorm.DB) error {
	return db.Migrator().DropTable(&entities.SavedAddress{})
}
