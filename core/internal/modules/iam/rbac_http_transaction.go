package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type rbacCommandAction func(
	context.Context, platformtx.TenantScope, uuid.UUID,
) error

type rbacSnapshotAction func(
	context.Context, platformtx.TenantScope, uuid.UUID,
) error

func (application *rbacHTTPApplication) inCommand(
	ctx context.Context,
	claims accesstoken.Claims,
	operationID string,
	action string,
	execute rbacCommandAction,
) error {
	if application == nil || execute == nil {
		return ErrRBACHTTPUnavailable
	}
	request, err := application.commandScopeRequest(ctx, claims, operationID)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorizeAdmin(
				scoped, scope, claims, action,
			)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		},
	)
	return rbacHTTPError(err)
}

func (application *rbacHTTPApplication) inSnapshot(
	ctx context.Context,
	evidence identityEvidenceSource,
	resource string,
	action string,
	execute rbacSnapshotAction,
) error {
	return application.inSnapshotAny(ctx, evidence, resource, []string{action}, execute)
}

func (application *rbacHTTPApplication) inSnapshotAny(
	ctx context.Context,
	evidence identityEvidenceSource,
	resource string,
	actions []string,
	execute rbacSnapshotAction,
) error {
	if application == nil || execute == nil {
		return ErrRBACHTTPUnavailable
	}
	if len(actions) == 0 {
		return ErrRBACHTTPForbidden
	}
	request, err := application.snapshotScopeRequest(ctx, evidence)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			for _, action := range actions {
				decision, authorizeErr := application.authorize(
					scoped, scope, evidence, resource, action,
				)
				if authorizeErr != nil {
					return authorizeErr
				}
				if !decision.Allowed() {
					continue
				}
				permissionID, valid := decision.PermissionID()
				if !valid || permissionID == uuid.Nil {
					return ErrRBACHTTPUnavailable
				}
				return execute(scoped, scope, permissionID)
			}
			return ErrRBACHTTPForbidden
		},
	)
	return rbacHTTPError(err)
}
