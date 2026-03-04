package controller

import (
	"errors"
	"fmt"
)

const (
	ActionEntityRead   = "entity.read"
	ActionEntityWrite  = "entity.write"
	ActionEntityDelete = "entity.delete"
	ActionEventRead    = "event.read"
	ActionEventWrite   = "event.write"
	ActionStreamRead   = "stream.read"
)

type Authorizer struct {
	matrix map[string]map[string]struct{}
}

func NewAuthorizer() *Authorizer {
	roles := map[string][]string{
		"player": {
			ActionEntityRead,
			ActionEventRead,
			ActionEventWrite,
			ActionStreamRead,
		},
		"moderator": {
			ActionEntityRead,
			ActionEntityWrite,
			ActionEventRead,
			ActionEventWrite,
			ActionStreamRead,
		},
		"developer": {
			ActionEntityRead,
			ActionEntityWrite,
			ActionEntityDelete,
			ActionEventRead,
			ActionEventWrite,
			ActionStreamRead,
		},
		"tenant_admin": {
			ActionEntityRead,
			ActionEntityWrite,
			ActionEntityDelete,
			ActionEventRead,
			ActionEventWrite,
			ActionStreamRead,
		},
		"platform_admin": {
			ActionEntityRead,
			ActionEntityWrite,
			ActionEntityDelete,
			ActionEventRead,
			ActionEventWrite,
			ActionStreamRead,
		},
	}

	matrix := make(map[string]map[string]struct{}, len(roles))
	for role, actions := range roles {
		actionSet := make(map[string]struct{}, len(actions))
		for _, action := range actions {
			actionSet[action] = struct{}{}
		}
		matrix[role] = actionSet
	}

	return &Authorizer{matrix: matrix}
}

func (a *Authorizer) Allow(role, action string) bool {
	actions, ok := a.matrix[role]
	if !ok {
		return false
	}
	_, ok = actions[action]
	return ok
}

func (a *Authorizer) Enforce(claims *Claims, action string) error {
	if claims == nil {
		return errors.New("missing claims")
	}
	if !a.Allow(claims.Role, action) {
		return fmt.Errorf("role %q is not allowed for %q", claims.Role, action)
	}
	return nil
}
