package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260823143000_create_sessions_table", UpCreateSessionsTable, DownCreateSessionsTable)
}

func UpCreateSessionsTable(db *gorm.DB) error {
	return db.AutoMigrate(&entities.Session{})
}

func DownCreateSessionsTable(db *gorm.DB) error {
	return db.Migrator().DropTable(&entities.Session{})
}
