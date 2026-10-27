package postgres

import "github.com/google/uuid"

type actorDocument struct {
	ID                   uuid.UUID `json:"id"`
	Status               string    `json:"status"`
	TokenVersion         int64     `json:"token_version"`
	AuthorizationVersion int64     `json:"authorization_version"`
}

type catalogDocument struct {
	ID        uuid.UUID `json:"id"`
	State     string    `json:"state"`
	Protected bool      `json:"protected"`
}

type bindingDocument struct {
	Path                string     `json:"path"`
	BindingID           uuid.UUID  `json:"binding_id"`
	BindingState        string     `json:"binding_state"`
	PermissionID        uuid.UUID  `json:"permission_id"`
	PermissionMatchKind string     `json:"permission_match_kind"`
	PermissionState     string     `json:"permission_state"`
	PermissionProtected bool       `json:"permission_protected"`
	PermissionResource  *uuid.UUID `json:"permission_resource"`
	PermissionAction    *uuid.UUID `json:"permission_action"`
	ResourceState       string     `json:"resource_state"`
	ActionState         string     `json:"action_state"`
	AssignmentID        *uuid.UUID `json:"assignment_id"`
	AssignmentState     string     `json:"assignment_state"`
	RoleID              *uuid.UUID `json:"role_id"`
	RoleStatus          string     `json:"role_status"`
	RoleScopeKind       string     `json:"role_scope_kind"`
	RoleProtected       bool       `json:"role_protected"`
	MembershipID        *uuid.UUID `json:"membership_id"`
	MembershipState     string     `json:"membership_state"`
	OrganizationID      *uuid.UUID `json:"organization_id"`
	OrganizationStatus  string     `json:"organization_status"`
	PolicyRoleID        *uuid.UUID `json:"policy_role_id"`
	PolicyPermissionID  *uuid.UUID `json:"policy_permission_id"`
	PolicyBindingID     *uuid.UUID `json:"policy_binding_id"`
	PolicyState         string     `json:"policy_state"`
}

type implicationDocument struct {
	ID                     uuid.UUID `json:"id"`
	Kind                   string    `json:"kind"`
	State                  string    `json:"state"`
	SourcePermissionID     uuid.UUID `json:"source_permission_id"`
	SourcePermissionState  string    `json:"source_permission_state"`
	SourceResourceID       uuid.UUID `json:"source_resource_id"`
	SourceActionID         uuid.UUID `json:"source_action_id"`
	SourceResourceState    string    `json:"source_resource_state"`
	SourceActionState      string    `json:"source_action_state"`
	ImpliedPermissionID    uuid.UUID `json:"implied_permission_id"`
	ImpliedPermissionState string    `json:"implied_permission_state"`
	ImpliedResourceID      uuid.UUID `json:"implied_resource_id"`
	ImpliedActionID        uuid.UUID `json:"implied_action_id"`
	ImpliedResourceState   string    `json:"implied_resource_state"`
	ImpliedActionState     string    `json:"implied_action_state"`
}
