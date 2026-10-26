package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type portalUserScopeAction func(
	context.Context, platformtx.TenantScope, uuid.UUID,
) error

func (application *portalUserHTTPApplication) inCommand(
	ctx context.Context, claims accesstoken.Claims, operation, action string,
	execute portalUserScopeAction,
) error {
	if application == nil || execute == nil {
		return ErrPortalUserHTTPUnavailable
	}
	request, err := application.portalUserCommandRequest(ctx, claims, operation)
	if err != nil {
		return ErrPortalUserHTTPUnavailable
	}
	err = application.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorize(scoped, scope, claims, action)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		})
	return mapPortalUserApplicationError(err)
}

func (application *portalUserHTTPApplication) portalUserCommandRequest(
	ctx context.Context, claims accesstoken.Claims, operation string,
) (platformtx.ScopeRequest, error) {
	if operation == "iam.portal_user.create" {
		return application.portalUserCreateRequest(ctx, claims)
	}
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(generations) {
		return platformtx.ScopeRequest{}, ErrPortalUserHTTPUnavailable
	}
	return platformtx.AdmitCommandScope(
		claims.Subject(), claims.Tenant(), operation, identityAuthorityGroup,
		generations["iam"], generations,
	)
}

func (application *portalUserHTTPApplication) portalUserCreateRequest(
	ctx context.Context, claims accesstoken.Claims,
) (platformtx.ScopeRequest, error) {
	names := []string{"billing", "forms", "iam", "system_resource_layout"}
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), names,
	)
	if err != nil || len(generations) != len(names) || generations["billing"] < 1 ||
		generations["billing"] != generations["forms"] || generations["iam"] < 1 ||
		generations["iam"] != generations["system_resource_layout"] {
		return platformtx.ScopeRequest{}, ErrPortalUserHTTPUnavailable
	}
	return platformtx.AdmitMultiGroupCommandScope(
		claims.Subject(), claims.Tenant(), "iam.portal_user.create",
		[]platformtx.CommandAuthorityGroup{
			{ID: "billing_forms", Generation: generations["billing"]},
			{ID: identityAuthorityGroup, Generation: generations["iam"]},
		}, []platformtx.CommandAuthorityDomain{
			{Name: "billing", Generation: generations["billing"]},
			{Name: "forms", Generation: generations["forms"]},
			{Name: "iam", Generation: generations["iam"]},
			{Name: "system_resource_layout", Generation: generations["system_resource_layout"]},
		}, []string{"forms", "iam"},
	)
}

func (application *portalUserHTTPApplication) inSnapshot(
	ctx context.Context, claims accesstoken.Claims, action string,
	execute portalUserScopeAction,
) error {
	generations, err := application.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(generations) {
		return ErrPortalUserHTTPUnavailable
	}
	request, err := queryapp.SnapshotScopeRequest(
		claims.Subject(), claims.Tenant(), map[string]int64{"iam": generations["iam"]},
	)
	if err != nil {
		return ErrPortalUserHTTPUnavailable
	}
	err = application.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			permissionID, authorizeErr := application.authorize(scoped, scope, claims, action)
			if authorizeErr != nil {
				return authorizeErr
			}
			return execute(scoped, scope, permissionID)
		})
	return mapPortalUserApplicationError(err)
}

func (application *portalUserHTTPApplication) authorize(
	ctx context.Context, scope platformtx.TenantScope,
	claims accesstoken.Claims, action string,
) (uuid.UUID, error) {
	request, err := authorization.NewRequest(
		"iam.users", action, claims.TokenVersion(), claims.AuthorizationVersion(),
	)
	if err != nil {
		return uuid.Nil, ErrPortalUserHTTPForbidden
	}
	decision, err := application.authorizer.Authorize(ctx, scope, request)
	if err != nil || decision.Validate() != nil {
		return uuid.Nil, ErrPortalUserHTTPUnavailable
	}
	permissionID, allowed := decision.PermissionID()
	if !decision.Allowed() || !allowed {
		return uuid.Nil, ErrPortalUserHTTPForbidden
	}
	return permissionID, nil
}

func mapPortalUserApplicationError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, portaluser.ErrInvalid):
		return ErrPortalUserHTTPInvalid
	case errors.Is(err, portaluser.ErrForbidden):
		return ErrPortalUserHTTPForbidden
	case errors.Is(err, portaluser.ErrNotFound):
		return ErrPortalUserHTTPNotFound
	case errors.Is(err, portaluser.ErrConflict):
		return ErrPortalUserHTTPConflict
	case errors.Is(err, ErrPortalUserHTTPInvalid), errors.Is(err, ErrPortalUserHTTPForbidden),
		errors.Is(err, ErrPortalUserHTTPNotFound), errors.Is(err, ErrPortalUserHTTPConflict):
		return err
	default:
		return ErrPortalUserHTTPUnavailable
	}
}
