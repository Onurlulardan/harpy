package serviceprincipal

import (
	"regexp"

	"github.com/google/uuid"
)

var resourceKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type ScopeKind string

const (
	ScopeTenant   ScopeKind = "tenant"
	ScopeResource ScopeKind = "resource"
)

type ResourceScope struct {
	Kind         ScopeKind
	Owner        string
	ResourceKind string
	ResourceID   uuid.UUID
}

func (scope ResourceScope) Validate() error {
	if !resourceKeyPattern.MatchString(scope.Owner) ||
		!resourceKeyPattern.MatchString(scope.ResourceKind) {
		return ErrInvalidGrant
	}
	switch scope.Kind {
	case ScopeTenant:
		if scope.ResourceID != uuid.Nil {
			return ErrInvalidGrant
		}
	case ScopeResource:
		if scope.ResourceID == uuid.Nil {
			return ErrInvalidGrant
		}
	default:
		return ErrInvalidGrant
	}
	return nil
}

func (scope ResourceScope) Equal(other ResourceScope) bool {
	return scope == other
}
