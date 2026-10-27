package iam

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

type PermissionRequestCreateItem struct {
	TargetKind   permissionrequest.TargetKind
	TargetID     uuid.UUID
	PermissionID uuid.UUID
}

type PermissionRequestCreate struct {
	Mode           permissionrequest.Mode
	ExpiresAt      time.Time
	Items          []PermissionRequestCreateItem
	IdempotencyKey string
}

type PermissionRequestList struct {
	Limit int
	After uuid.UUID
}

type PermissionRequestCatalogList struct {
	Kind   permissionrequest.CatalogKind
	Search string
	Limit  int
	After  uuid.UUID
}

type PermissionRequestMissingCatalogList struct {
	PermissionRequestCatalogList
	ActionKey string
}

type PermissionRequestCatalogResourceView struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	IsSystem    bool    `json:"is_system"`
	Kind        string  `json:"kind"`
	FormSlug    *string `json:"form_slug,omitempty"`
}

type PermissionRequestCatalogActionView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

type PermissionRequestCatalogResourcePageView struct {
	Items       []PermissionRequestCatalogResourceView
	HasMore     bool
	NextAfterID *string
}

type PermissionRequestCatalogActionPageView struct {
	Items       []PermissionRequestCatalogActionView
	HasMore     bool
	NextAfterID *string
}

type PermissionRequestReview struct {
	ID               uuid.UUID
	ExpectedRevision int64
	ExpectedState    permissionrequest.State
	IdempotencyKey   string
}

type PermissionRequestDecision struct {
	ID               uuid.UUID
	ExpectedRevision int64
	ExpectedState    permissionrequest.State
	ReviewTaskID     uuid.UUID
	IdempotencyKey   string
}

type PermissionRequestTerminal struct {
	ID               uuid.UUID
	ExpectedRevision int64
	ExpectedState    permissionrequest.State
}

type PermissionRequestItemView struct {
	ID                  string  `json:"id"`
	Ordinal             int16   `json:"ordinal"`
	TargetKind          string  `json:"target_kind"`
	TargetID            string  `json:"target_id"`
	PermissionID        string  `json:"permission_id"`
	PermissionProtected bool    `json:"permission_protected"`
	PlannedBindingID    string  `json:"planned_binding_id"`
	GrantedBindingID    *string `json:"granted_binding_id,omitempty"`
	State               string  `json:"state"`
	Revision            int64   `json:"revision"`
}

type PermissionRequestGroupView struct {
	ID                  string                      `json:"id"`
	RequesterID         string                      `json:"requester_id"`
	Mode                string                      `json:"mode"`
	State               string                      `json:"state"`
	ReviewerID          *string                     `json:"reviewer_id,omitempty"`
	ReviewTaskID        *string                     `json:"review_task_id,omitempty"`
	Revision            int64                       `json:"revision"`
	ExpiresAt           string                      `json:"expires_at"`
	TerminalAt          string                      `json:"terminal_at,omitempty"`
	CreatedAt           string                      `json:"created_at"`
	UpdatedAt           string                      `json:"updated_at"`
	StepUpPurposeSHA256 string                      `json:"step_up_purpose_sha256"`
	Items               []PermissionRequestItemView `json:"items"`
}

type PermissionRequestRefView struct {
	ID                  string `json:"id"`
	State               string `json:"state"`
	Revision            int64  `json:"revision"`
	StepUpPurposeSHA256 string `json:"step_up_purpose_sha256"`
}

type PermissionRequestPageView struct {
	Items       []PermissionRequestGroupView `json:"items"`
	HasMore     bool                         `json:"has_more"`
	NextAfterID *string                      `json:"next_after_id,omitempty"`
}
