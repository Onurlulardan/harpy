package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *rbacHTTPApplication) Check(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACCheck,
) (RBACDecisionView, error) {
	if !rbacCheckValid(input) {
		return RBACDecisionView{}, ErrRBACHTTPInvalid
	}
	request, err := application.snapshotScopeRequest(ctx, claims)
	if err != nil {
		return RBACDecisionView{}, err
	}
	var view RBACDecisionView
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := application.authorize(
				scoped, scope, claims, input.ResourceKey, input.ActionKey,
			)
			if authorizeErr != nil {
				return authorizeErr
			}
			view = RBACDecisionView{
				Allowed: decision.Allowed(), Reason: string(decision.Reason()),
				AuthorizationClock:   decision.AuthorizationClock(),
				AuthorizationVersion: decision.AuthorizationVersion(),
			}
			return nil
		},
	)
	if err != nil {
		return RBACDecisionView{}, rbacHTTPError(err)
	}
	return view, nil
}
