package script

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/AdiCahyaSaputra/go-ojol/backend/auth/pkg/constants"
	"gorm.io/gorm"
)

const casbinSeedDir = "./database/seeders/json"

type casbinRuleSeed struct {
	Ptype string `json:"ptype"`
	V0    string `json:"v0"`
	V1    string `json:"v1"`
	V2    string `json:"v2,omitempty"`
}

type CasbinSeedScript struct {
	db *gorm.DB
}

func NewCasbinSeedScript(db *gorm.DB) *CasbinSeedScript {
	return &CasbinSeedScript{db: db}
}

func (s *CasbinSeedScript) Run() error {
	if err := os.MkdirAll(casbinSeedDir, 0o755); err != nil {
		return err
	}

	if err := writeCasbinJSON("casbin_policies.json", casbinPolicySeeds()); err != nil {
		return err
	}

	return writeCasbinJSON("casbin_grouping.json", casbinGroupingSeeds())
}

func casbinPolicySeeds() []casbinRuleSeed {
	return []casbinRuleSeed{
		{Ptype: "p", V0: constants.ENUM_ROLE_ADMIN, V1: constants.ENUM_RESOURCE_USER, V2: constants.ENUM_ACTION_READ},
		{Ptype: "p", V0: constants.ENUM_ROLE_ADMIN, V1: constants.ENUM_RESOURCE_USER, V2: constants.ENUM_ACTION_UPDATE},
		{Ptype: "p", V0: constants.ENUM_ROLE_ADMIN, V1: constants.ENUM_RESOURCE_USER, V2: constants.ENUM_ACTION_DELETE},
		{Ptype: "p", V0: constants.ENUM_ROLE_CUSTOMER, V1: constants.ENUM_RESOURCE_USER, V2: constants.ENUM_ACTION_READ},
		{Ptype: "p", V0: constants.ENUM_ROLE_CUSTOMER, V1: constants.ENUM_RESOURCE_USER, V2: constants.ENUM_ACTION_UPDATE},
		{Ptype: "p", V0: constants.ENUM_ROLE_DRIVER, V1: constants.ENUM_RESOURCE_USER, V2: constants.ENUM_ACTION_READ},
		{Ptype: "p", V0: constants.ENUM_ROLE_DRIVER, V1: constants.ENUM_RESOURCE_USER, V2: constants.ENUM_ACTION_UPDATE},
		{Ptype: "p", V0: constants.ENUM_ROLE_ADMIN, V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_CREATE},
		{Ptype: "p", V0: constants.ENUM_ROLE_ADMIN, V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_READ},
		{Ptype: "p", V0: constants.ENUM_ROLE_ADMIN, V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_UPDATE},
		{Ptype: "p", V0: constants.ENUM_ROLE_ADMIN, V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_DELETE},
		{Ptype: "p", V0: constants.ENUM_ROLE_CUSTOMER, V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_CREATE},
		{Ptype: "p", V0: constants.ENUM_ROLE_CUSTOMER, V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_READ},
		{Ptype: "p", V0: constants.ENUM_ROLE_DRIVER, V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_READ},
		{Ptype: "p", V0: constants.ENUM_ROLE_DRIVER, V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_UPDATE},
	}
}

func casbinGroupingSeeds() []casbinRuleSeed {
	return []casbinRuleSeed{
		{Ptype: "g", V0: "adm.adics@gmail.com", V1: constants.ENUM_ROLE_ADMIN},
		{Ptype: "g", V0: "cst.adics@gmail.com", V1: constants.ENUM_ROLE_CUSTOMER},
		{Ptype: "g", V0: "drv.adics@gmail.com", V1: constants.ENUM_ROLE_DRIVER},
	}
}

func writeCasbinJSON(filename string, rules []casbinRuleSeed) error {
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(casbinSeedDir, filename), data, 0o644)
}
