package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260816162215_vehicle", UpVehicle, DownVehicle)
}

func UpVehicle(db *gorm.DB) error {
	db.Exec("create type vehicle_type as enum('car', 'motorcycle');")

	return db.AutoMigrate(&entities.Vehicle{})
}

func DownVehicle(db *gorm.DB) error {
	db.Exec("drop type vehicle_type;")

	return db.Migrator().DropTable(&entities.Vehicle{})
}
