package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) ListResources(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACList,
) (RBACResourcePage, error) {
	return application.listResourcesWithEvidence(ctx, claims, input, []string{"read"})
}

func (application *rbacHTTPApplication) listResourcesWithEvidence(
	ctx context.Context,
	evidence identityEvidenceSource,
	input RBACList,
	actions []string,
) (RBACResourcePage, error) {
	if !rbacListValid(input) || input.AfterID != uuid.Nil {
		return RBACResourcePage{}, ErrRBACHTTPInvalid
	}
	var values []rbacdomain.Resource
	err := application.inSnapshotAny(
		ctx, evidence, "iam.permissions", actions,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var listErr error
			values, listErr = application.runtime.resources.List(
				scoped, scope, rbacdomain.ListQuery{
					State: input.State, Limit: input.Limit + 1,
					Evidence: rbacQueryEvidence(evidence, permissionID),
				},
			)
			return listErr
		},
	)
	if err != nil {
		return RBACResourcePage{}, err
	}
	hasMore := len(values) > input.Limit
	if hasMore {
		values = values[:input.Limit]
	}
	page := RBACResourcePage{Items: make([]RBACResourceView, 0, len(values)), HasMore: hasMore}
	for _, value := range values {
		view := rbacResourceView(value)
		if !validRBACResourceView(view) {
			return RBACResourcePage{}, ErrRBACHTTPUnavailable
		}
		page.Items = append(page.Items, view)
	}
	return page, nil
}
