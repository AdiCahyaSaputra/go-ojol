package migrations

import (
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260829140000_transaction_trip_location_fields", UpTransactionTripLocationFields, DownTransactionTripLocationFields)
}

func UpTransactionTripLocationFields(db *gorm.DB) error {
	if db.Migrator().HasColumn(&entities.Transaction{}, "last_lat_long") {
		if err := db.Exec(`ALTER TABLE transactions RENAME COLUMN last_lat_long TO driver_last_lat_long`).Error; err != nil {
			return err
		}
	}

	if !db.Migrator().HasColumn(&entities.Transaction{}, "customer_last_lat_long") {
		if err := db.Exec(`ALTER TABLE transactions ADD COLUMN customer_last_lat_long varchar(40)[]`).Error; err != nil {
			return err
		}
	}

	if !db.Migrator().HasColumn(&entities.Transaction{}, "paid_at") {
		if err := db.Exec(`ALTER TABLE transactions ADD COLUMN paid_at timestamptz`).Error; err != nil {
			return err
		}
	}

	if err := db.Exec(`UPDATE transactions SET customer_last_lat_long = pickup_lat_long WHERE customer_last_lat_long IS NULL`).Error; err != nil {
		return err
	}

	return db.AutoMigrate(&entities.Transaction{})
}

func DownTransactionTripLocationFields(db *gorm.DB) error {
	if db.Migrator().HasColumn(&entities.Transaction{}, "paid_at") {
		if err := db.Migrator().DropColumn(&entities.Transaction{}, "paid_at"); err != nil {
			return err
		}
	}

	if db.Migrator().HasColumn(&entities.Transaction{}, "customer_last_lat_long") {
		if err := db.Migrator().DropColumn(&entities.Transaction{}, "customer_last_lat_long"); err != nil {
			return err
		}
	}

	if db.Migrator().HasColumn(&entities.Transaction{}, "driver_last_lat_long") {
		if err := db.Exec(`ALTER TABLE transactions RENAME COLUMN driver_last_lat_long TO last_lat_long`).Error; err != nil {
			return err
		}
	}

	return nil
}
