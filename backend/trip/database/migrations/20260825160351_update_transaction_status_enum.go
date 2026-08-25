package migrations

import (
	"fmt"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() {
	database.RegisterMigration("20260825160351_update_transaction_status_enum", UpUpdateTransactionStatusEnum, DownUpdateTransactionStatusEnum)
}

var valueBeforePair = gin.H{
	"offered":        "on_the_way",
	"accepted_offer": "on_the_way",
	"rejected_offer": "on_the_way",
}

var valueAfterPair = gin.H{
	"expired": "completed",
}

func UpUpdateTransactionStatusEnum(db *gorm.DB) error {
	for value, beforeValue := range valueBeforePair {
		query := fmt.Sprintf("alter type transaction_status add value if not exists '%s' before '%s'", value, beforeValue)
		db.Exec(query)
	}

	for value, afterValue := range valueAfterPair {
		query := fmt.Sprintf("alter type transaction_status add value if not exists '%s' after '%s'", value, afterValue)
		db.Exec(query)
	}

	return db.AutoMigrate(&entities.Transaction{})
}

func DownUpdateTransactionStatusEnum(db *gorm.DB) error {
	for value := range valueBeforePair {
		query := fmt.Sprintf("alter type transaction_status drop value '%s'", value)
		db.Exec(query)
	}

	for value := range valueAfterPair {
		query := fmt.Sprintf("alter type transaction_status drop value '%s'", value)
		db.Exec(query)
	}

	return nil
}
