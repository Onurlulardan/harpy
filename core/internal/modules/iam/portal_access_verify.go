package iam

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *PortalAccessRuntime) VerifyPortalAccess(
	ctx context.Context, request PortalAccessRequest,
) (PortalAccessDecision, error) {
	if !runtime.Ready() || ctx == nil || ctx.Err() != nil {
		return PortalAccessDecision{}, ErrPortalAccessUnavailable
	}
	wire, valid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.clock().UTC().Truncate(time.Second)
	if !valid || now.IsZero() || now.Unix() <= 0 {
		return PortalAccessDecision{}, ErrPortalAccessAuthentication
	}
	claims, err := runtime.security.accessVerifier.Verify(ctx, wire, now)
	bindingID, hasBinding := claims.PortalBinding()
	if err != nil || claims.AccessMode() != accesstoken.AccessModePortal ||
		claims.Client() != accesstoken.ClientPortal || !hasBinding ||
		!tenantAccessHeadersMatch(ctx, runtime.pool, request.TenantIDValues,
			request.TenantSlugValues, claims.Tenant()) {
		return PortalAccessDecision{}, ErrPortalAccessAuthentication
	}
	requestScope, err := iamHTTPSnapshotScopeRequest(ctx, runtime.pool, claims)
	if err != nil {
		return PortalAccessDecision{}, ErrPortalAccessUnavailable
	}
	var binding portaluser.AccessBinding
	err = runtime.coordinator.InTenant(ctx, requestScope,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var verifyErr error
			binding, verifyErr = runtime.users.VerifyAccess(scoped, scope,
				claims.Subject(), bindingID, claims.TokenVersion(), claims.AuthorizationVersion())
			return verifyErr
		})
	if err != nil {
		return PortalAccessDecision{}, mapPortalAccessError(err)
	}
	decision, err := newPortalAccessDecision(binding)
	if err != nil {
		return PortalAccessDecision{}, ErrPortalAccessUnavailable
	}
	return decision, nil
}

func mapPortalAccessError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, portaluser.ErrForbidden) || errors.Is(err, portaluser.ErrInvalid) ||
		errors.Is(err, portaluser.ErrNotFound) {
		return ErrPortalAccessForbidden
	}
	return ErrPortalAccessUnavailable
}
