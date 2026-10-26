package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type lifecycleTenantAction func(context.Context, platformtx.TenantScope, uuid.UUID) error

func (application *lifecycleHTTPApplication) inCommand(
	ctx context.Context, claims accesstoken.Claims, operation, action string,
	execute lifecycleTenantAction,
) error {
	if application == nil || execute == nil || !validLifecycleClaims(claims) {
		return ErrLifecycleHTTPInvalid
	}
	request, err := application.commandScope(ctx, claims, operation)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorize(scoped, scope, claims, action)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		})
	return mapLifecycleHTTPError(err)
}

func (application *lifecycleHTTPApplication) inSnapshot(
	ctx context.Context, claims accesstoken.Claims, action string,
	execute lifecycleTenantAction,
) error {
	if application == nil || execute == nil || !validLifecycleClaims(claims) {
		return ErrLifecycleHTTPInvalid
	}
	request, err := application.snapshotScope(ctx, claims)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorize(scoped, scope, claims, action)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		})
	return mapLifecycleHTTPError(err)
}

func validLifecycleClaims(claims accesstoken.Claims) bool {
	return claims.Subject() != uuid.Nil && claims.Session() != uuid.Nil &&
		claims.Tenant() != uuid.Nil && claims.TokenVersion() > 0 &&
		claims.AuthorizationVersion() > 0
}
