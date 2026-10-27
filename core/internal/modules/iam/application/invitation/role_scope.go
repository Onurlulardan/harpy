package invitation

import "github.com/google/uuid"

type RoleScopeKind string

const (
	RoleScopeTenant       RoleScopeKind = "tenant"
	RoleScopeOrganization RoleScopeKind = "organization"
)

type RoleScope struct {
	Kind           RoleScopeKind
	ID             uuid.UUID
	OrganizationID uuid.UUID
}

func (scope RoleScope) Validate() error {
	if scope.ID == uuid.Nil {
		return ErrInvalid
	}
	if scope.Kind == RoleScopeTenant && scope.OrganizationID == uuid.Nil {
		return nil
	}
	if scope.Kind == RoleScopeOrganization &&
		scope.OrganizationID != uuid.Nil && scope.OrganizationID == scope.ID {
		return nil
	}
	return ErrInvalid
}
