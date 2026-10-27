package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) ListPermissions(
	ctx context.Context,
	claims accesstoken.Claims,
	input RBACList,
) (RBACPermissionPage, error) {
	return application.listPermissionsWithEvidence(ctx, claims, input, []string{"read"})
}

func (application *rbacHTTPApplication) listPermissionsWithEvidence(
	ctx context.Context,
	evidence identityEvidenceSource,
	input RBACList,
	actions []string,
) (RBACPermissionPage, error) {
	if !rbacListValid(input) {
		return RBACPermissionPage{}, ErrRBACHTTPInvalid
	}
	page := RBACPermissionPage{}
	err := application.inSnapshotAny(
		ctx, evidence, "iam.permissions", actions,
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			values, listErr := application.runtime.assignments.List(
				scoped, scope, rbacdomain.AssignmentListQuery{
					State: input.State, Limit: input.Limit + 1,
					AfterID:  rbacdomain.ID(input.AfterID),
					Evidence: rbacQueryEvidence(evidence, permissionID),
				},
			)
			if listErr != nil {
				return listErr
			}
			page.HasMore = len(values) > input.Limit
			if page.HasMore {
				values = values[:input.Limit]
			}
			page.Items = make([]RBACPermissionView, 0, len(values))
			for _, value := range values {
				view, viewErr := application.assignmentView(
					scoped, scope, evidence, permissionID, value,
				)
				if viewErr != nil || !validRBACPermissionView(view) {
					return rbacdomain.ErrStorage
				}
				page.Items = append(page.Items, view)
			}
			if page.HasMore && len(page.Items) > 0 {
				next := page.Items[len(page.Items)-1].ID
				page.NextAfterID = &next
			}
			return nil
		},
	)
	if err != nil {
		return RBACPermissionPage{}, err
	}
	return page, nil
}
