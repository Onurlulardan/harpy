package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *CatalogAccessRuntime) Verify(
	ctx context.Context, request CatalogAccessRequest,
) (CatalogAccessDecision, error) {
	if runtime == nil || runtime.shared == nil || ctx == nil || ctx.Err() != nil {
		return CatalogAccessDecision{}, ErrCatalogAccessUnavailable
	}
	target, targetValid := resolveCatalogAccess(request.Permission, request.Operation)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.shared.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return CatalogAccessDecision{}, ErrCatalogAccessAuthentication
	}
	claims, err := runtime.shared.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!runtime.tenantHeadersMatch(ctx, request, claims.Tenant()) {
		return CatalogAccessDecision{}, ErrCatalogAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return CatalogAccessDecision{}, ErrCatalogAccessForbidden
	}
	stepUpAt, stepUpValid := catalogStepUpTime(claims, request, now)
	if !stepUpValid {
		return CatalogAccessDecision{}, ErrCatalogAccessForbidden
	}
	authorizationDecision, err := runtime.authorize(ctx, claims, target)
	if err != nil {
		return CatalogAccessDecision{}, err
	}
	permissionID, available := authorizationDecision.PermissionID()
	if !available {
		return CatalogAccessDecision{}, ErrCatalogAccessUnavailable
	}
	stepUpPurpose := [32]byte{}
	if !stepUpAt.IsZero() {
		stepUpPurpose = catalogStepUpPurpose(
			claims.Tenant(), request.Operation, request.PurposeTarget,
		)
	}
	evidenceSHA256 := catalogAuthorizationEvidenceSHA256(
		claims.Subject(), claims.Tenant(), claims.TokenVersion(),
		claims.AuthorizationVersion(), authorizationDecision.AuthorizationClock(),
		permissionID, request.Permission, request.Operation, target,
		stepUpAt, stepUpPurpose,
	)
	decision, err := NewCatalogAccessDecisionWithEvidence(
		claims.Subject(), claims.Tenant(),
		claims.TokenVersion(), claims.AuthorizationVersion(),
		authorizationDecision.AuthorizationClock(), stepUpAt, permissionID,
		evidenceSHA256, stepUpPurpose,
	)
	if err != nil {
		return CatalogAccessDecision{}, ErrCatalogAccessUnavailable
	}
	return decision, nil
}
