package auth

import (
	"fmt"

	"github.com/casbin/casbin/v2"
)

type Enforcer interface {
	Enforce(sub, obj, act string) (bool, error)
}

type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

func NewCasbinEnforcer(modelPath, policyPath string) (*CasbinEnforcer, error) {
	enforcer, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize casbin enforcer: %w", err)
	}

	return &CasbinEnforcer{
		enforcer: enforcer,
	}, nil
}

func (c *CasbinEnforcer) Enforce(sub, obj, act string) (bool, error) {
	return c.enforcer.Enforce(sub, obj, act)
}
