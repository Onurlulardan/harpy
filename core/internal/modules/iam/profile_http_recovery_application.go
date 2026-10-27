package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *profileHTTPApplication) RecoverBackfill(
	ctx context.Context,
	claims accesstoken.Claims,
	input ProfileBackfillRecoveryInput,
) (ProfileBackfillRecoveryView, error) {
	if input.IdempotencyKey == "" || !input.Resource.Valid() ||
		input.ParentPlanID == uuid.Nil || application == nil ||
		application.recoveries == nil {
		return ProfileBackfillRecoveryView{}, ErrProfileHTTPInvalid
	}
	requestSHA, err := profileRecoveryRequestSHA(input)
	if err != nil {
		return ProfileBackfillRecoveryView{}, err
	}
	operationID := identityOperationID(
		claims, "iam.profile.recover", input.IdempotencyKey,
	)
	var view ProfileBackfillRecoveryView
	err = application.inProfileCommand(
		ctx, claims, "iam.profile.publish", input.Resource,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			result, recoverErr := application.recoveries.Recover(
				scoped, scope, input.Resource, input.ParentPlanID,
				profileMutationEvidence(claims, permissionID, operationID, requestSHA),
			)
			if recoverErr == nil {
				view = profileRecoveryView(result)
			}
			return recoverErr
		},
	)
	return view, err
}
