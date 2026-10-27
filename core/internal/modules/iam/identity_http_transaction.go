package iam

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identityScopeAction func(
	context.Context, platformtx.TenantScope, uuid.UUID,
) error

func (application *identityHTTPApplication) inCommand(
	ctx context.Context,
	evidence identityEvidenceSource,
	authorityOperation string,
	resource IdentityResource,
	action string,
	execute identityScopeAction,
) error {
	if application == nil || execute == nil {
		return ErrIdentityHTTPUnavailable
	}
	request, err := application.commandScopeRequest(ctx, evidence, authorityOperation)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorize(
				scoped, scope, evidence, resource, action,
			)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		},
	)
	return identityApplicationError(err)
}

func (application *identityHTTPApplication) inSnapshot(
	ctx context.Context,
	evidence identityEvidenceSource,
	authorityOperation string,
	resource IdentityResource,
	execute identityScopeAction,
) error {
	return application.inSnapshotWithAction(
		ctx, evidence, authorityOperation, resource, "read", execute,
	)
}

func (application *identityHTTPApplication) inSnapshotWithAction(
	ctx context.Context,
	evidence identityEvidenceSource,
	authorityOperation string,
	resource IdentityResource,
	action string,
	execute identityScopeAction,
) error {
	if application == nil || execute == nil {
		return ErrIdentityHTTPUnavailable
	}
	request, err := application.snapshotScopeRequest(ctx, evidence, authorityOperation)
	if err != nil {
		return err
	}
	err = application.coordinator.InTenant(
		ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorize(
				scoped, scope, evidence, resource, action,
			)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		},
	)
	return identityApplicationError(err)
}
