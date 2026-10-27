package iam

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (runtime *FormsRecordAccessRuntime) Verify(
	ctx context.Context,
	request FormsRecordAccessRequest,
) (FormsRecordAccessDecision, error) {
	if runtime == nil || ctx == nil || ctx.Err() != nil {
		return FormsRecordAccessDecision{}, ErrFormsRecordAccessUnavailable
	}
	target, valid := resolveFormsRecordAccess(request.Action, request.Operation)
	wire, bearerValid := formsRecordBearer(request.AuthorizationValues)
	now := runtime.clock().UTC().Truncate(time.Second)
	if !valid || !bearerValid || now.IsZero() || now.Unix() <= 0 {
		return FormsRecordAccessDecision{}, ErrFormsRecordAccessAuthentication
	}
	claims, err := runtime.security.accessVerifier.Verify(ctx, wire, now)
	if err != nil || claims.AccessMode() != accesstoken.AccessModeOperator ||
		!identityHTTPClientAllowed(claims.Client()) ||
		!runtime.tenantHeadersMatch(ctx, request, claims.Tenant()) {
		return FormsRecordAccessDecision{}, ErrFormsRecordAccessAuthentication
	}
	if _, portal := claims.PortalBinding(); portal {
		return FormsRecordAccessDecision{}, ErrFormsRecordAccessForbidden
	}
	authorizationDecision, err := runtime.authorizeFormsRecord(ctx, claims, target)
	if err != nil {
		return FormsRecordAccessDecision{}, err
	}
	decision, err := NewFormsRecordAccessDecision(
		claims.Subject(), claims.Tenant(), claims.TokenVersion(),
		claims.AuthorizationVersion(), authorizationDecision.AuthorizationClock(),
	)
	if err != nil {
		return FormsRecordAccessDecision{}, ErrFormsRecordAccessUnavailable
	}
	return decision, nil
}
