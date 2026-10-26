package iam

import (
	"encoding/json"

	"github.com/google/uuid"
)

type IdentityResource string

const (
	IdentityUsers         IdentityResource = "users"
	IdentityRoles         IdentityResource = "roles"
	IdentityOrganizations IdentityResource = "organizations"
)

type IdentityMutation struct {
	ID               uuid.UUID
	ExpectedRevision int64
	IdempotencyKey   string
}

type IdentityBulkDelete struct {
	IdempotencyKey string
	Items          []IdentityMutation
}

type IdentityStateMutation struct {
	IdentityMutation
	State string
}

type IdentityUserCreate struct {
	IdempotencyKey   string
	Email            string
	Password         []byte
	FirstName        string
	LastName         string
	Phone            string
	OrganizationID   uuid.UUID
	RoleIDs          []uuid.UUID
	CustomAttributes json.RawMessage
}

func (input *IdentityUserCreate) Destroy() {
	if input == nil {
		return
	}
	clear(input.Password)
	input.Password = nil
}

type IdentityUserUpdate struct {
	IdentityMutation
	FirstName        *string
	LastName         *string
	Phone            *string
	OrganizationID   *uuid.UUID
	RoleIDs          *[]uuid.UUID
	CustomAttributes json.RawMessage
	LegacyState      string
}

type IdentityRoleCreate struct {
	IdempotencyKey   string
	Name             string
	Description      string
	ScopeKind        string
	ScopeID          uuid.UUID
	OrganizationID   uuid.UUID
	Default          bool
	CustomAttributes json.RawMessage
}

type IdentityRoleUpdate struct {
	IdentityMutation
	Name             *string
	Description      *string
	Default          *bool
	CustomAttributes json.RawMessage
}

type IdentityOrganizationCreate struct {
	IdempotencyKey   string
	Name             string
	Slug             string
	ParentID         uuid.UUID
	OwnerIDs         []uuid.UUID
	CustomAttributes json.RawMessage
}

type IdentityOrganizationUpdate struct {
	IdentityMutation
	Name             *string
	Slug             *string
	ParentID         *uuid.UUID
	OwnerIDs         *[]uuid.UUID
	CustomAttributes json.RawMessage
}
