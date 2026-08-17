package database

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	manager := NewMigrationManager(db)
	return manager.Run()
}
