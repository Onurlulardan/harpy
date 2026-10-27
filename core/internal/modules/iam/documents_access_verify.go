package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *DocumentsAccessRuntime) Verify(ctx context.Context,
	request DocumentsAccessRequest,
) (DocumentsAccessDecision, error) {
	if runtime == nil || runtime.shared == nil || ctx == nil || ctx.Err() != nil {
		return DocumentsAccessDecision{}, ErrDocumentsAccessUnavailable
	}
	target, targetValid := resolveDocumentsAccess(request.Permission, request.Operation)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.shared.clock().UTC().Truncate(time.Second)
	if !targetValid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return DocumentsAccessDecision{}, ErrDocumentsAccessAuthentication
	}
	claims, err := runtime.shared.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!runtime.tenantHeadersMatch(ctx, request, claims.Tenant()) {
		return DocumentsAccessDecision{}, ErrDocumentsAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return DocumentsAccessDecision{}, ErrDocumentsAccessForbidden
	}
	stepUpAt, stepUpValid := documentsStepUpTime(claims, request, now)
	if !stepUpValid {
		return DocumentsAccessDecision{}, ErrDocumentsAccessForbidden
	}
	authorizationDecision, err := runtime.authorize(ctx, claims, target)
	if err != nil {
		return DocumentsAccessDecision{}, err
	}
	decision, err := NewDocumentsAccessDecision(claims.Subject(), claims.Tenant(),
		claims.TokenVersion(), claims.AuthorizationVersion(),
		authorizationDecision.AuthorizationClock(), stepUpAt)
	if err != nil {
		return DocumentsAccessDecision{}, ErrDocumentsAccessUnavailable
	}
	return decision, nil
}
