package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) ListActions(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACList,
) (RBACActionPage, error) {
	return application.listActionsWithEvidence(ctx, claims, input, []string{"read"})
}

func (application *rbacHTTPApplication) listActionsWithEvidence(
	ctx context.Context,
	evidence identityEvidenceSource,
	input RBACList,
	actions []string,
) (RBACActionPage, error) {
	if !rbacListValid(input) || input.AfterID != uuid.Nil {
		return RBACActionPage{}, ErrRBACHTTPInvalid
	}
	var values []rbacdomain.Action
	err := application.inSnapshotAny(
		ctx, evidence, "iam.permissions", actions,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var listErr error
			values, listErr = application.runtime.actions.List(
				scoped, scope, rbacdomain.ListQuery{
					State: input.State, Limit: input.Limit + 1,
					Evidence: rbacQueryEvidence(evidence, permissionID),
				},
			)
			return listErr
		},
	)
	if err != nil {
		return RBACActionPage{}, err
	}
	hasMore := len(values) > input.Limit
	if hasMore {
		values = values[:input.Limit]
	}
	page := RBACActionPage{Items: make([]RBACActionView, 0, len(values)), HasMore: hasMore}
	for _, value := range values {
		view := rbacActionView(value)
		if !validRBACActionView(view) {
			return RBACActionPage{}, ErrRBACHTTPUnavailable
		}
		page.Items = append(page.Items, view)
	}
	return page, nil
}
