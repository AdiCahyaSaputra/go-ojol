package casbin

import (
	"embed"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"gorm.io/gorm"
)

//go:embed rbac_model.conf
var rbacModelFS embed.FS

type Enforcer interface {
	Enforce(rvals ...any) (bool, error)
	LoadPolicy() error
}

func NewEnforcer(db *gorm.DB) (Enforcer, error) {
	modelText, err := rbacModelFS.ReadFile("rbac_model.conf")
	if err != nil {
		return nil, err
	}

	m, err := model.NewModelFromString(string(modelText))
	if err != nil {
		return nil, err
	}

	e, err := casbin.NewEnforcer(m, NewAdapter(db))
	if err != nil {
		return nil, err
	}

	e.EnableAutoSave(false)
	return e, nil
}
