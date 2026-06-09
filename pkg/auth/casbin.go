package auth

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type Enforcer interface {
	Enforce(sub, obj, act string) (bool, error)
	AddPolicy(sub, obj, act string) (bool, error)
	RemovePolicy(sub, obj, act string) (bool, error)
	GetPolicy() ([][]string, error)
	GetGroupingPolicy() ([][]string, error)
	AddGroupingPolicy(params ...string) (bool, error)
	RemoveGroupingPolicy(params ...string) (bool, error)
}

type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

func NewCasbinEnforcer(modelPath string, db *gorm.DB) (*CasbinEnforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize casbin gorm adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize casbin enforcer: %w", err)
	}

	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to load casbin policies: %w", err)
	}

	return &CasbinEnforcer{
		enforcer: enforcer,
	}, nil
}

func (c *CasbinEnforcer) Enforce(sub, obj, act string) (bool, error) {
	return c.enforcer.Enforce(sub, obj, act)
}

func (c *CasbinEnforcer) AddPolicy(sub, obj, act string) (bool, error) {
	return c.enforcer.AddPolicy(sub, obj, act)
}

func (c *CasbinEnforcer) RemovePolicy(sub, obj, act string) (bool, error) {
	return c.enforcer.RemovePolicy(sub, obj, act)
}

func (c *CasbinEnforcer) GetPolicy() ([][]string, error) {
	return c.enforcer.GetPolicy()
}

func (c *CasbinEnforcer) GetGroupingPolicy() ([][]string, error) {
	return c.enforcer.GetGroupingPolicy()
}

func (c *CasbinEnforcer) AddGroupingPolicy(params ...string) (bool, error) {
	return c.enforcer.AddGroupingPolicy(params)
}

func (c *CasbinEnforcer) RemoveGroupingPolicy(params ...string) (bool, error) {
	return c.enforcer.RemoveGroupingPolicy(params)
}
