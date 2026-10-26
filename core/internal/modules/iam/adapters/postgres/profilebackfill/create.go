package profilebackfill

import (
	"context"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func (value *store) Create(
	ctx context.Context,
	scope tenantScope,
	prepared profilebackfillapp.PreparedPlan,
) (profilebackfillapp.Result, error) {
	evidence := prepared.Command.Evidence
	if ctx == nil || ctx.Err() != nil || !scopeUsable(ctx, scope) ||
		scope.ActorID() != evidence.ActorID || len(prepared.Document) == 0 {
		return profilebackfillapp.Result{}, profilebackfillapp.ErrInvalid
	}
	bound, err := value.bind(scope)
	if err != nil {
		return profilebackfillapp.Result{}, err
	}
	var result profilebackfillapp.Result
	err = bound.database.QueryRow(
		ctx, createSQL, scope.TenantID(), evidence.OperationID,
		evidence.RequestSHA256[:], prepared.Document, prepared.PlanSHA256[:],
		evidence.ActorID, evidence.ManagePermissionID,
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	).Scan(&result.PlanID, &result.Revision, &result.Replayed)
	if err != nil {
		return profilebackfillapp.Result{}, err
	}
	receipt, found, err := scanReceipt(
		bound.database.QueryRow(
			ctx, readReceiptSQL, evidence.OperationID, evidence.ActorID,
			evidence.ManagePermissionID, evidence.ExpectedTokenVersion,
			evidence.ExpectedAuthorizationVersion,
		),
	)
	if err != nil || !found || receipt.CommandKey != "iam.profile_backfill.plan_create" ||
		receipt.RequestSHA256 != evidence.RequestSHA256 ||
		receipt.ResultID != result.PlanID || receipt.Revision != result.Revision {
		return profilebackfillapp.Result{}, profilebackfillapp.ErrUnavailable
	}
	result.AuthorizationClock = receipt.AuthorizationClock
	return result, nil
}
