package iam

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type IdentityUserView struct {
	ID               uuid.UUID       `json:"id"`
	Email            string          `json:"email"`
	FirstName        string          `json:"first_name"`
	LastName         string          `json:"last_name"`
	Phone            string          `json:"phone,omitempty"`
	EmailVerified    bool            `json:"email_verified"`
	AccountType      string          `json:"account_type,omitempty"`
	Status           string          `json:"status"`
	RoleIDs          []uuid.UUID     `json:"role_ids"`
	OrganizationIDs  []uuid.UUID     `json:"organization_ids"`
	OrganizationID   *uuid.UUID      `json:"organization_id,omitempty"`
	CustomAttributes json.RawMessage `json:"custom_attributes,omitempty"`
	Revision         int64           `json:"revision"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type IdentityRoleView struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	ScopeKind        string          `json:"scope_kind"`
	ScopeID          uuid.UUID       `json:"scope_id"`
	OrganizationID   uuid.UUID       `json:"organization_id,omitempty"`
	Default          bool            `json:"is_default"`
	Protected        bool            `json:"is_protected"`
	Status           string          `json:"status"`
	CustomAttributes json.RawMessage `json:"custom_attributes,omitempty"`
	Revision         int64           `json:"revision"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type IdentityOrganizationView struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	Slug             string          `json:"slug"`
	Status           string          `json:"status"`
	ParentID         uuid.UUID       `json:"parent_id,omitempty"`
	OwnerIDs         []uuid.UUID     `json:"owner_ids"`
	OwnerID          *uuid.UUID      `json:"owner_id,omitempty"`
	CustomAttributes json.RawMessage `json:"custom_attributes,omitempty"`
	Revision         int64           `json:"revision"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type IdentityQueryView struct {
	Items         []json.RawMessage
	HasMore       bool
	NextCursor    string
	NextTransport string
	ExactCount    *int64
}
