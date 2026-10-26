package authorization

import "github.com/google/uuid"

type SubjectPath string

const (
	SubjectUser         SubjectPath = "user"
	SubjectRole         SubjectPath = "role"
	SubjectOrganization SubjectPath = "organization"
)

type BindingFact struct {
	Path                SubjectPath
	BindingID           uuid.UUID
	BindingState        string
	PermissionID        uuid.UUID
	PermissionMatchKind string
	PermissionState     string
	PermissionProtected bool
	PermissionResource  *uuid.UUID
	PermissionAction    *uuid.UUID
	ResourceState       string
	ActionState         string
	AssignmentID        *uuid.UUID
	AssignmentState     string
	RoleID              *uuid.UUID
	RoleStatus          string
	RoleScopeKind       string
	RoleProtected       bool
	MembershipID        *uuid.UUID
	MembershipState     string
	OrganizationID      *uuid.UUID
	OrganizationStatus  string
	PolicyRoleID        *uuid.UUID
	PolicyPermissionID  *uuid.UUID
	PolicyBindingID     *uuid.UUID
	PolicyState         string
}

func (fact BindingFact) valid() bool {
	if fact.BindingID == uuid.Nil || fact.PermissionID == uuid.Nil ||
		!validBindingState(fact.BindingState) ||
		!validCatalogState(fact.PermissionState) || !fact.validPermissionShape() {
		return false
	}
	switch fact.Path {
	case SubjectUser:
		return fact.validUserShape()
	case SubjectRole:
		return fact.validRoleShape()
	case SubjectOrganization:
		return fact.validOrganizationShape()
	default:
		return false
	}
}

func (fact BindingFact) validPermissionShape() bool {
	switch fact.PermissionMatchKind {
	case "exact":
		return validUUIDPointer(fact.PermissionResource) &&
			validUUIDPointer(fact.PermissionAction) &&
			validCatalogState(fact.ResourceState) && validCatalogState(fact.ActionState)
	case "resource_all_actions":
		return fact.PermissionProtected && validUUIDPointer(fact.PermissionResource) &&
			fact.PermissionAction == nil && validCatalogState(fact.ResourceState) &&
			fact.ActionState == ""
	case "tenant_all":
		return fact.PermissionProtected && fact.PermissionResource == nil &&
			fact.PermissionAction == nil && fact.ResourceState == "" &&
			fact.ActionState == ""
	default:
		return false
	}
}

func validUUIDPointer(value *uuid.UUID) bool {
	return value != nil && *value != uuid.Nil
}

func validBindingState(value string) bool {
	return value == "active" || value == "revoked"
}
