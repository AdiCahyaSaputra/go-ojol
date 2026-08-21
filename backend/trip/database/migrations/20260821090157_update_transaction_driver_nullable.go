package migrations

import (
	"fmt"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260821090157_update_transaction_driver_nullable", UpUpdateTransactionDriverNullable, DownUpdateTransactionDriverNullable)
}

func UpUpdateTransactionDriverNullable(db *gorm.DB) error {
	shouldNullableFIDs := []string{
		"customer_id",
		"driver_id",
		"vehicle_id",
	}

	for _, field := range shouldNullableFIDs {
		db.Exec(fmt.Sprintf("ALTER TABLE transactions ALTER COLUMN %s DROP NOT NULL", field))
	}

	return nil
}

func DownUpdateTransactionDriverNullable(db *gorm.DB) error {
	shouldNullableFIDs := []string{
		"customer_id",
		"driver_id",
		"vehicle_id",
	}

	for _, field := range shouldNullableFIDs {
		db.Exec(fmt.Sprintf("ALTER TABLE transactions ALTER COLUMN %s SET NOT NULL", field))
	}

	return nil
}
