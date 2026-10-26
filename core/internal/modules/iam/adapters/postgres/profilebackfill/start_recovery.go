package profilebackfill

import (
	"context"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func (value *store) StartRecovery(
	ctx context.Context,
	scope tenantScope,
	prepared profilebackfillapp.PreparedRecoveryPlan,
) (profilebackfillapp.RecoveryResult, error) {
	evidence := prepared.Command.Evidence
	plan := prepared.Command.Plan
	if value == nil || evidence.Validate() != nil || plan.Validate() != nil ||
		len(prepared.Document) < 2 || prepared.PlanSHA256 == [32]byte{} ||
		!scopeUsable(ctx, scope) || scope.ActorID() != evidence.ActorID {
		return profilebackfillapp.RecoveryResult{}, profilebackfillapp.ErrInvalid
	}
	bound, err := value.bind(scope)
	if err != nil {
		return profilebackfillapp.RecoveryResult{}, err
	}
	var result profilebackfillapp.RecoveryResult
	err = bound.database.QueryRow(
		ctx, createRecoverySQL, scope.TenantID(), evidence.OperationID,
		evidence.RequestSHA256[:], prepared.Document, prepared.PlanSHA256[:],
		evidence.ActorID, evidence.ManagePermissionID,
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	).Scan(
		&result.PlanID, &result.Revision, &result.AuthorizationClock,
		&result.State, &result.Replayed,
	)
	if err != nil {
		return profilebackfillapp.RecoveryResult{}, recoveryStoreError(err)
	}
	receipt, found, err := readRecoveryReceipt(ctx, bound, evidence)
	if err != nil || !found || receipt.RequestSHA256 != evidence.RequestSHA256 ||
		receipt.PlanID != result.PlanID || receipt.PlanID != plan.ID() ||
		receipt.RecoveryOfPlanID != plan.RecoveryOfPlanID() ||
		receipt.RecoveryRootPlanID != plan.RecoveryRootPlanID() ||
		receipt.RecoverySequence != plan.RecoverySequence() ||
		receipt.Revision != result.Revision || receipt.State != result.State ||
		receipt.AuthorizationClock != result.AuthorizationClock {
		return profilebackfillapp.RecoveryResult{},
			profilebackfillapp.ErrUnavailable
	}
	return result, nil
}
