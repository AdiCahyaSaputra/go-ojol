package casbin

import (
	"fmt"
	"testing"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEnforcer_EnforceByEmailResourceAction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE casbin_rules (
			id TEXT PRIMARY KEY,
			ptype TEXT NOT NULL,
			v0 TEXT,
			v1 TEXT,
			v2 TEXT,
			v3 TEXT,
			v4 TEXT,
			v5 TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error)

	rules := []entities.CasbinRule{
		{Ptype: entities.CasbinRulePtypePerm, V0: "admin", V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_CREATE},
		{Ptype: entities.CasbinRulePtypePerm, V0: "admin", V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_READ},
		{Ptype: entities.CasbinRulePtypePerm, V0: "admin", V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_UPDATE},
		{Ptype: entities.CasbinRulePtypePerm, V0: "admin", V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_DELETE},
		{Ptype: entities.CasbinRulePtypePerm, V0: "customer", V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_CREATE},
		{Ptype: entities.CasbinRulePtypePerm, V0: "customer", V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_READ},
		{Ptype: entities.CasbinRulePtypePerm, V0: "driver", V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_READ},
		{Ptype: entities.CasbinRulePtypePerm, V0: "driver", V1: constants.ENUM_RESOURCE_TRIP, V2: constants.ENUM_ACTION_UPDATE},
		{Ptype: entities.CasbinRulePtypeRole, V0: "admin@example.com", V1: "admin"},
		{Ptype: entities.CasbinRulePtypeRole, V0: "ada@example.com", V1: "customer"},
		{Ptype: entities.CasbinRulePtypeRole, V0: "drv@example.com", V1: "driver"},
	}
	require.NoError(t, db.Create(&rules).Error)

	enforcer, err := NewEnforcer(db)
	require.NoError(t, err)

	allowed, err := enforcer.Enforce("ada@example.com", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_CREATE)
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = enforcer.Enforce("ada@example.com", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_READ)
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = enforcer.Enforce("ada@example.com", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_DELETE)
	require.NoError(t, err)
	assert.False(t, allowed)

	allowed, err = enforcer.Enforce("drv@example.com", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_CREATE)
	require.NoError(t, err)
	assert.False(t, allowed)

	allowed, err = enforcer.Enforce("drv@example.com", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_UPDATE)
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = enforcer.Enforce("admin@example.com", constants.ENUM_RESOURCE_TRIP, constants.ENUM_ACTION_DELETE)
	require.NoError(t, err)
	assert.True(t, allowed)
}
