package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbac/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *rbacHTTPApplication) readResource(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (rbacdomain.Resource, error) {
	var result rbacdomain.Resource
	err := application.inSnapshot(
		ctx, claims, "iam.permissions", "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			result, readErr = application.runtime.resources.Get(
				scoped, scope, rbacdomain.GetQuery{
					ID: rbacdomain.ID(id), Evidence: rbacQueryEvidence(claims, permissionID),
				},
			)
			return readErr
		},
	)
	return result, err
}

func (application *rbacHTTPApplication) readAction(
	ctx context.Context,
	claims accesstoken.Claims,
	id uuid.UUID,
) (rbacdomain.Action, error) {
	var result rbacdomain.Action
	err := application.inSnapshot(
		ctx, claims, "iam.permissions", "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			result, readErr = application.runtime.actions.Get(
				scoped, scope, rbacdomain.GetQuery{
					ID: rbacdomain.ID(id), Evidence: rbacQueryEvidence(claims, permissionID),
				},
			)
			return readErr
		},
	)
	return result, err
}

func (application *rbacHTTPApplication) readPermissionCatalog(
	ctx context.Context,
	claims accesstoken.Claims,
	resourceID uuid.UUID,
	actionIDs []uuid.UUID,
) (RBACResourceView, []RBACActionView, error) {
	resource, err := application.readResource(ctx, claims, resourceID)
	if err != nil {
		return RBACResourceView{}, nil, err
	}
	actions := make([]RBACActionView, 0, len(actionIDs))
	for _, actionID := range identitySortedIDs(actionIDs) {
		action, actionErr := application.readAction(ctx, claims, actionID)
		if actionErr != nil {
			return RBACResourceView{}, nil, actionErr
		}
		actions = append(actions, rbacActionView(action))
	}
	return rbacResourceView(resource), actions, nil
}
