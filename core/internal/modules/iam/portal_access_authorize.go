package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *PortalAccessRuntime) AuthorizePortalAction(
	ctx context.Context, access PortalAccessDecision, resourceSlug, action string,
) error {
	_, err := runtime.AuthorizePortalActionGrant(ctx, access, resourceSlug, action)
	return err
}

func (runtime *PortalAccessRuntime) AuthorizePortalActionGrant(
	ctx context.Context, access PortalAccessDecision, resourceSlug, action string,
) (PortalActionGrant, error) {
	if !runtime.Ready() || ctx == nil || ctx.Err() != nil || !access.Valid() {
		return PortalActionGrant{}, ErrPortalAccessUnavailable
	}
	request, err := authorization.NewRequest(
		resourceSlug, action, access.TokenVersion(), access.AuthorizationVersion(),
	)
	if err != nil {
		return PortalActionGrant{}, ErrPortalAccessForbidden
	}
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, access.TenantID(), []string{"iam", "system_resource_layout"},
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(generations) {
		return PortalActionGrant{}, ErrPortalAccessUnavailable
	}
	scopeRequest, err := queryapp.SnapshotScopeRequest(access.PrincipalID(), access.TenantID(),
		map[string]int64{"iam": generations["iam"]})
	if err != nil {
		return PortalActionGrant{}, ErrPortalAccessUnavailable
	}
	var grant PortalActionGrant
	err = runtime.coordinator.InTenant(ctx, scopeRequest,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			current, verifyErr := runtime.users.VerifyAccess(scoped, scope,
				access.PrincipalID(), access.BindingID(), access.TokenVersion(),
				access.AuthorizationVersion())
			if verifyErr != nil || !samePortalActionBinding(current, access) {
				return ErrPortalAccessForbidden
			}
			decision, authorizeErr := runtime.authorizer.Authorize(scoped, scope, request)
			if authorizeErr != nil || decision.Validate() != nil {
				return ErrPortalAccessUnavailable
			}
			if !decision.Allowed() {
				return ErrPortalAccessForbidden
			}
			permissionID, ok := decision.PermissionID()
			if !ok || decision.AuthorizationVersion() != access.AuthorizationVersion() ||
				decision.AuthorizationClock() < access.AuthorizationClock() {
				return ErrPortalAccessUnavailable
			}
			grant, verifyErr = newPortalActionGrant(access, permissionID,
				decision.AuthorizationVersion(), decision.AuthorizationClock(),
				resourceSlug, action)
			return verifyErr
		})
	if err == ErrPortalAccessForbidden {
		return PortalActionGrant{}, err
	}
	if err != nil {
		return PortalActionGrant{}, ErrPortalAccessUnavailable
	}
	if !grant.ValidFor(access, resourceSlug, action) {
		return PortalActionGrant{}, ErrPortalAccessUnavailable
	}
	return grant, nil
}
