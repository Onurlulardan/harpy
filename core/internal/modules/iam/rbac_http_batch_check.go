package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *rbacHTTPApplication) BatchCheck(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACBatchCheck,
) (RBACBatchDecisionView, error) {
	if !rbacBatchCheckValid(input) {
		return RBACBatchDecisionView{}, ErrRBACHTTPInvalid
	}
	request, err := application.snapshotScopeRequest(ctx, claims)
	if err != nil {
		return RBACBatchDecisionView{}, err
	}
	view := RBACBatchDecisionView{Results: make(map[string]bool, len(input.Checks))}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			for _, check := range input.Checks {
				decision, authorizeErr := application.authorize(
					scoped, scope, claims, check.ResourceKey, check.ActionKey,
				)
				if authorizeErr != nil {
					return authorizeErr
				}
				view.Results[check.ResourceKey+":"+check.ActionKey] = decision.Allowed()
			}
			return nil
		},
	)
	if err != nil {
		return RBACBatchDecisionView{}, rbacHTTPError(err)
	}
	return view, nil
}
