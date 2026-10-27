package iam

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

type lifecyclePortalSetupView struct {
	BindingID        uuid.UUID  `json:"binding_id"`
	UserID           uuid.UUID  `json:"user_id"`
	SetupState       string     `json:"setup_state"`
	ApprovalState    string     `json:"approval_state"`
	Revision         int64      `json:"revision"`
	SetupCompletedAt *time.Time `json:"setup_completed_at,omitempty"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	DecisionAt       *time.Time `json:"decision_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func lifecyclePortalSetupResponse(
	value lifecycle.PortalSetup,
) (lifecyclePortalSetupView, bool) {
	if value.Validate() != nil {
		return lifecyclePortalSetupView{}, false
	}
	result := lifecyclePortalSetupView{
		BindingID: value.BindingID, UserID: value.UserID,
		SetupState: value.SetupState, ApprovalState: value.ApprovalState,
		Revision: value.Revision, UpdatedAt: value.UpdatedAt.UTC(),
	}
	result.SetupCompletedAt = lifecycleOptionalTime(value.SetupCompletedAt)
	result.SubmittedAt = lifecycleOptionalTime(value.SubmittedAt)
	result.DecisionAt = lifecycleOptionalTime(value.DecisionAt)
	return result, true
}

func lifecycleOptionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	utc := value.UTC()
	return &utc
}
