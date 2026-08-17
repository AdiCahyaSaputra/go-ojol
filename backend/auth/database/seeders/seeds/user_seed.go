package seeds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userSeed struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	VehicleID   string `json:"vehicle_id"`
}

func ListUserSeeder(db *gorm.DB) error {
	jsonFile, err := os.Open("./database/seeders/json/users.json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var listUser []userSeed
	if err := json.Unmarshal(jsonData, &listUser); err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entities.User{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entities.User{}); err != nil {
			return err
		}
	}

	for _, data := range listUser {
		if err := seedUser(db, data); err != nil {
			return err
		}
	}

	return nil
}

func seedUser(db *gorm.DB, data userSeed) error {
	var existing entities.User
	err := db.Where(&entities.User{Email: data.Email}).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		return nil
	}

	user := entities.User{
		Email:    data.Email,
		Password: data.Password,
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}

	switch data.Role {
	case constants.ENUM_ROLE_ADMIN:
		return nil
	case constants.ENUM_ROLE_CUSTOMER:
		return seedCustomer(db, user.ID, data)
	case constants.ENUM_ROLE_DRIVER:
		return seedDriver(db, user.ID, data)
	default:
		return fmt.Errorf("unsupported seed role %q for %s", data.Role, data.Email)
	}
}

func seedCustomer(db *gorm.DB, userID uuid.UUID, data userSeed) error {
	if data.Name == "" || data.PhoneNumber == "" {
		return fmt.Errorf("customer seed %s requires name and phone_number", data.Email)
	}

	customer := entities.Customer{
		UserID:      userID,
		Name:        data.Name,
		PhoneNumber: data.PhoneNumber,
	}
	return db.Create(&customer).Error
}

func seedDriver(db *gorm.DB, userID uuid.UUID, data userSeed) error {
	if data.Name == "" || data.PhoneNumber == "" || data.Address == "" || data.VehicleID == "" {
		return fmt.Errorf("driver seed %s requires name, phone_number, address, and vehicle_id", data.Email)
	}

	vehicleID, err := uuid.Parse(data.VehicleID)
	if err != nil {
		return fmt.Errorf("driver seed %s has invalid vehicle_id: %w", data.Email, err)
	}

	driver := entities.Driver{
		UserID:      userID,
		VehicleID:   vehicleID,
		Name:        data.Name,
		PhoneNumber: data.PhoneNumber,
		Address:     data.Address,
	}
	return db.Create(&driver).Error
}
