package casbin

import (
	"fmt"

	"github.com/AdiCahyaSaputra/go-ojol/backend/trip/database/entities"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(db *gorm.DB) persist.Adapter {
	return &Adapter{db: db}
}

func (a *Adapter) LoadPolicy(m model.Model) error {
	var rules []entities.CasbinRule
	if err := a.db.Find(&rules).Error; err != nil {
		return err
	}

	for _, rule := range rules {
		if err := persist.LoadPolicyArray(ruleToArgs(rule), m); err != nil {
			return err
		}
	}

	return nil
}

func (a *Adapter) SavePolicy(m model.Model) error {
	if err := a.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entities.CasbinRule{}).Error; err != nil {
		return err
	}

	var lines []entities.CasbinRule

	for ptype, ast := range m["p"] {
		for _, rule := range ast.Policy {
			lines = append(lines, policyToRule(ptype, rule))
		}
	}
	for ptype, ast := range m["g"] {
		for _, rule := range ast.Policy {
			lines = append(lines, policyToRule(ptype, rule))
		}
	}

	if len(lines) == 0 {
		return nil
	}

	return a.db.Create(&lines).Error
}

func (a *Adapter) AddPolicy(_ string, ptype string, rule []string) error {
	line := policyToRule(ptype, rule)
	return a.db.Create(&line).Error
}

func (a *Adapter) RemovePolicy(_ string, ptype string, rule []string) error {
	line := policyToRule(ptype, rule)
	return a.db.Where(
		"ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ? AND v5 = ?",
		line.Ptype, line.V0, line.V1, line.V2, line.V3, line.V4, line.V5,
	).Delete(&entities.CasbinRule{}).Error
}

func (a *Adapter) RemoveFilteredPolicy(_ string, ptype string, fieldIndex int, fieldValues ...string) error {
	query := a.db.Where("ptype = ?", ptype)

	for i, value := range fieldValues {
		if value == "" {
			continue
		}
		query = query.Where(fmt.Sprintf("v%d = ?", fieldIndex+i), value)
	}

	return query.Delete(&entities.CasbinRule{}).Error
}

func policyToRule(ptype string, rule []string) entities.CasbinRule {
	line := entities.CasbinRule{Ptype: entities.CasbinRulePtype(ptype)}
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}
	return line
}

func ruleToArgs(rule entities.CasbinRule) []string {
	args := []string{string(rule.Ptype), rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5}
	i := len(args) - 1
	for i >= 0 && args[i] == "" {
		i--
	}
	return args[:i+1]
}
