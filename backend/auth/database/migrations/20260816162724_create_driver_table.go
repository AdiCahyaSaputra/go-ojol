package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260816162724_drivers", UpDrivers, DownDrivers)
}

func UpDrivers(db *gorm.DB) error {
	return db.AutoMigrate(&entities.Driver{})
}

func DownDrivers(db *gorm.DB) error {
	return db.Migrator().DropTable(&entities.Driver{})
}
