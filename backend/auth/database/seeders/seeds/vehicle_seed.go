package seeds

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"gorm.io/gorm"
)

func ListVehicleSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/seeders/json/vehicles.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var listVehicle []entities.Vehicle
	if err := json.Unmarshal(jsonData, &listVehicle); err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entities.Vehicle{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entities.Vehicle{}); err != nil {
			return err
		}
	}

	for _, data := range listVehicle {
		var existing entities.Vehicle
		err := db.Where("license_number = ?", data.LicenseNumber).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			continue
		}

		if err := db.Create(&data).Error; err != nil {
			return err
		}
	}

	return nil
}
