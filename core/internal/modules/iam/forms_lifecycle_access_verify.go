package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *FormsLifecycleAccessRuntime) Verify(
	ctx context.Context,
	request FormsLifecycleAccessRequest,
) (FormsLifecycleAccessDecision, error) {
	if runtime == nil || runtime.shared == nil || ctx == nil || ctx.Err() != nil {
		return FormsLifecycleAccessDecision{}, ErrFormsLifecycleAccessUnavailable
	}
	target, targetValid := resolveFormsLifecycleAccess(request.Permission, request.Operation)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.shared.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return FormsLifecycleAccessDecision{}, ErrFormsLifecycleAccessAuthentication
	}
	claims, err := runtime.shared.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!runtime.tenantHeadersMatch(ctx, request, claims.Tenant()) {
		return FormsLifecycleAccessDecision{}, ErrFormsLifecycleAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return FormsLifecycleAccessDecision{}, ErrFormsLifecycleAccessForbidden
	}
	authorizationDecision, err := runtime.authorize(ctx, claims, target)
	if err != nil {
		return FormsLifecycleAccessDecision{}, err
	}
	decision, err := NewFormsLifecycleAccessDecision(claims.Subject(), claims.Tenant(),
		claims.TokenVersion(), claims.AuthorizationVersion(),
		authorizationDecision.AuthorizationClock())
	if err != nil {
		return FormsLifecycleAccessDecision{}, ErrFormsLifecycleAccessUnavailable
	}
	return decision, nil
}
