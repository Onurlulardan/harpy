package iam

import (
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

type ProfileBackfillRecoveryInput struct {
	IdempotencyKey string
	Resource       profileapp.ResourceKind
	ParentPlanID   uuid.UUID
}

type ProfileBackfillRecoveryView struct {
	PlanID             uuid.UUID                        `json:"plan_id"`
	RecoveryOfPlanID   uuid.UUID                        `json:"recovery_of_plan_id"`
	RecoveryRootPlanID uuid.UUID                        `json:"recovery_root_plan_id"`
	RecoverySequence   int64                            `json:"recovery_sequence"`
	Revision           int64                            `json:"revision"`
	AuthorizationClock int64                            `json:"authorization_clock"`
	State              profilebackfillapp.RecoveryState `json:"state"`
	Replayed           bool                             `json:"replayed"`
}

func profileRecoveryRequestSHA(
	input ProfileBackfillRecoveryInput,
) ([32]byte, error) {
	digest, err := identityRequestSHA(struct {
		Resource     profileapp.ResourceKind
		ParentPlanID uuid.UUID
	}{input.Resource, input.ParentPlanID})
	if err != nil {
		return [32]byte{}, ErrProfileHTTPInvalid
	}
	return digest, nil
}

func profileRecoveryView(
	result ProfileBackfillRecoveryResult,
) ProfileBackfillRecoveryView {
	return ProfileBackfillRecoveryView{
		PlanID: result.PlanID, RecoveryOfPlanID: result.RecoveryOfPlanID,
		RecoveryRootPlanID: result.RecoveryRootPlanID,
		RecoverySequence:   result.RecoverySequence, Revision: result.Revision,
		AuthorizationClock: result.AuthorizationClock,
		State:              result.State, Replayed: result.Replayed,
	}
}
