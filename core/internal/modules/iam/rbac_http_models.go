package iam

import (
	rbac "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	"github.com/google/uuid"
)

type RBACMutation struct {
	ID               uuid.UUID
	ExpectedRevision int64
	IdempotencyKey   string
}

type RBACList struct {
	State   string
	Limit   int
	AfterID uuid.UUID
}

type RBACCatalogCreate struct {
	Key            string
	DisplayName    string
	Description    string
	IdempotencyKey string
}

type RBACCatalogUpdate struct {
	RBACMutation
	Key         string
	DisplayName string
	Description string
}

type RBACPermissionCreate struct {
	ResourceID     uuid.UUID
	Target         rbac.AssignmentTarget
	TargetID       uuid.UUID
	ActionIDs      []uuid.UUID
	IdempotencyKey string
}

type RBACPermissionUpdate struct {
	RBACMutation
	ResourceID uuid.UUID
	Target     rbac.AssignmentTarget
	TargetID   uuid.UUID
	ActionIDs  []uuid.UUID
}

type RBACCheck struct {
	ResourceKey string
	ActionKey   string
}

type RBACBatchCheck struct {
	Checks []RBACCheck
}

type RBACEffectiveRead struct {
	Target   rbac.AssignmentTarget
	TargetID uuid.UUID
	Limit    int
}
