package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type profileScopeAction func(
	context.Context, platformtx.TenantScope, uuid.UUID,
) error

func (application *profileHTTPApplication) inProfileSnapshot(
	ctx context.Context,
	claims accesstoken.Claims,
	resource profileapp.ResourceKind,
	execute profileScopeAction,
) error {
	if application == nil || execute == nil {
		return ErrProfileHTTPUnavailable
	}
	request, err := application.snapshotScopeRequest(ctx, claims)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorize(
				scoped, scope, claims, resource, "read",
			)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		},
	)
	return profileApplicationError(err)
}

func (application *profileHTTPApplication) inProfileCommand(
	ctx context.Context,
	claims accesstoken.Claims,
	operation string,
	resource profileapp.ResourceKind,
	execute profileScopeAction,
) error {
	if application == nil || execute == nil {
		return ErrProfileHTTPUnavailable
	}
	request, err := application.commandScopeRequest(ctx, claims, operation)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorize(
				scoped, scope, claims, resource, "manage",
			)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		},
	)
	return profileApplicationError(err)
}
