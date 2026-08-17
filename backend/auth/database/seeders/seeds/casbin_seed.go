package seeds

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/database/entities"
	"gorm.io/gorm"
)

type casbinRuleSeed struct {
	Ptype string `json:"ptype"`
	V0    string `json:"v0"`
	V1    string `json:"v1"`
	V2    string `json:"v2"`
}

func ListCasbinSeeder(db *gorm.DB) error {
	files, err := filepath.Glob("./database/seeders/json/casbin_*.json")
	if err != nil {
		return err
	}

	hasTable := db.Migrator().HasTable(&entities.CasbinRule{})
	if !hasTable {
		if err := db.Migrator().CreateTable(&entities.CasbinRule{}); err != nil {
			return err
		}
	}

	for _, file := range files {
		if err := seedCasbinFile(db, file); err != nil {
			return err
		}
	}

	return nil
}

func seedCasbinFile(db *gorm.DB, path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var rules []casbinRuleSeed
	if err := json.Unmarshal(jsonData, &rules); err != nil {
		return err
	}

	for _, data := range rules {
		var existing entities.CasbinRule
		err := db.Where(
			"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
			data.Ptype, data.V0, data.V1, data.V2,
		).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			continue
		}

		rule := entities.CasbinRule{
			Ptype: entities.CasbinRulePtype(data.Ptype),
			V0:    data.V0,
			V1:    data.V1,
			V2:    data.V2,
		}
		if err := db.Create(&rule).Error; err != nil {
			return err
		}
	}

	return nil
}
