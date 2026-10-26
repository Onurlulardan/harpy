package iam

import (
	"context"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *ProfileBackfillRecoveryRuntime) Recover(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource profileapp.ResourceKind,
	parentID uuid.UUID,
	evidence profileapp.MutationEvidence,
) (ProfileBackfillRecoveryResult, error) {
	if runtime == nil || runtime.source == nil || runtime.starter == nil ||
		runtime.clock == nil || !resource.Valid() || parentID == uuid.Nil ||
		evidence.Validate() != nil || scope.ActorID() != evidence.ActorID {
		return ProfileBackfillRecoveryResult{}, profilebackfillapp.ErrInvalid
	}
	sharedEvidence := profileBackfillEvidence(evidence)
	if replay, found, err := runtime.replay(
		ctx, scope, parentID, sharedEvidence,
	); err != nil || found {
		return replay, err
	}
	contract, err := runtime.source.InspectParent(
		ctx, scope, parentID, sharedEvidence,
	)
	if err != nil {
		return ProfileBackfillRecoveryResult{}, err
	}
	if contract.ExistingChildPlanID != uuid.Nil {
		replay, found, replayErr := runtime.replay(
			ctx, scope, parentID, sharedEvidence,
		)
		if replayErr != nil {
			return ProfileBackfillRecoveryResult{}, replayErr
		}
		if found {
			return replay, nil
		}
		return ProfileBackfillRecoveryResult{}, profileapp.ErrConflict
	}
	if contract.Source.Resource() != resource || contract.Target.Resource() != resource {
		return ProfileBackfillRecoveryResult{}, profileapp.ErrConflict
	}
	plan, err := runtime.plan(ctx, scope, contract, sharedEvidence)
	if err != nil {
		return ProfileBackfillRecoveryResult{}, err
	}
	result, err := runtime.starter.Start(
		ctx, scope, profilebackfillapp.RecoveryCreateCommand{
			Plan: plan, Evidence: sharedEvidence,
		},
	)
	if err != nil {
		return ProfileBackfillRecoveryResult{}, err
	}
	return recoveryResultFromPlan(plan, result), nil
}

func (runtime *ProfileBackfillRecoveryRuntime) replay(
	ctx context.Context,
	scope platformtx.TenantScope,
	parentID uuid.UUID,
	evidence profilebackfillapp.Evidence,
) (ProfileBackfillRecoveryResult, bool, error) {
	receipt, found, err := runtime.source.LookupReceipt(ctx, scope, evidence)
	if err != nil || !found {
		return ProfileBackfillRecoveryResult{}, false, err
	}
	if receipt.RequestSHA256 != evidence.RequestSHA256 ||
		receipt.RecoveryOfPlanID != parentID {
		return ProfileBackfillRecoveryResult{}, false, profileapp.ErrConflict
	}
	return recoveryResultFromReceipt(receipt), true, nil
}

func (runtime *ProfileBackfillRecoveryRuntime) recoveryRunAt() (time.Time, error) {
	runAt := runtime.clock().UTC()
	if runAt.IsZero() {
		return time.Time{}, profilebackfillapp.ErrUnavailable
	}
	return runAt, nil
}
