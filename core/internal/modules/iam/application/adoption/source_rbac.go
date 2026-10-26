package adoption

import "github.com/google/uuid"

type SourceResource struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Key         string
	DisplayName string
	System      bool
}

type SourceAction struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Key         string
	DisplayName string
	System      bool
}

type SourcePermission struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	ResourceID     uuid.UUID
	Target         string
	UserID         *uuid.UUID
	RoleID         *uuid.UUID
	OrganizationID *uuid.UUID
	ActionIDs      []uuid.UUID
}

type SourcePermissionRequest struct {
	ID          uuid.UUID
	RequesterID uuid.UUID
	TargetID    uuid.UUID
	ResourceID  uuid.UUID
	ActionID    uuid.UUID
	State       string
	GroupID     *uuid.UUID
}
