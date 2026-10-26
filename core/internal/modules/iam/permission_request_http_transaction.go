package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type permissionRequestScopeAction func(context.Context, platformtx.TenantScope) error

func (application *permissionRequestHTTPApplication) inCommand(
	ctx context.Context,
	claims accesstoken.Claims,
	operationID string,
	action string,
	execute permissionRequestScopeAction,
) error {
	if application == nil || execute == nil {
		return ErrPermissionRequestHTTPUnavailable
	}
	request, err := application.commandScopeRequest(ctx, claims, operationID)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			if action != "" {
				if authorizeErr := application.authorize(scoped, scope, claims, action); authorizeErr != nil {
					return authorizeErr
				}
			}
			return execute(scoped, scope)
		},
	)
	return permissionRequestHTTPError(err)
}

func (application *permissionRequestHTTPApplication) inSnapshot(
	ctx context.Context,
	evidence identityEvidenceSource,
	execute permissionRequestScopeAction,
) error {
	if application == nil || execute == nil {
		return ErrPermissionRequestHTTPUnavailable
	}
	request, err := application.snapshotScopeRequest(ctx, evidence)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(ctx, request, execute)
	return permissionRequestHTTPError(err)
}

func (application *permissionRequestHTTPApplication) inSnapshotAnyAction(
	ctx context.Context,
	evidence identityEvidenceSource,
	actions []string,
	execute permissionRequestScopeAction,
) error {
	if application == nil || execute == nil || len(actions) == 0 {
		return ErrPermissionRequestHTTPUnavailable
	}
	request, err := application.snapshotScopeRequest(ctx, evidence)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			for _, action := range actions {
				authorizeErr := application.authorize(scoped, scope, evidence, action)
				if authorizeErr == nil {
					return execute(scoped, scope)
				}
				if authorizeErr != ErrPermissionRequestHTTPForbidden {
					return authorizeErr
				}
			}
			return ErrPermissionRequestHTTPForbidden
		},
	)
	return permissionRequestHTTPError(err)
}
