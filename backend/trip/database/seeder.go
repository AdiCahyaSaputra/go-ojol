package database

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/seeders/seeds"
	"gorm.io/gorm"
)

func Seeder(db *gorm.DB) error {
	if err := seeds.ListUserSeeder(db); err != nil {
		return err
	}

	return nil
}
