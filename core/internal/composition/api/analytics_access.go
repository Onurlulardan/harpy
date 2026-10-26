package api

import (
	"context"
	"errors"
	"fmt"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type analyticsIAMAccess interface {
	Verify(context.Context, iammodule.AnalyticsAccessRequest) (
		iammodule.AnalyticsAccessDecision, error,
	)
}

type analyticsAccessVerifier struct{ iam analyticsIAMAccess }

func newAnalyticsAccessVerifier(iam analyticsIAMAccess) (*analyticsAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("Analytics IAM access is required")
	}
	return &analyticsAccessVerifier{iam: iam}, nil
}

func (verifier *analyticsAccessVerifier) Verify(
	ctx context.Context, request analyticsmodule.HTTPAccessRequest,
) (analyticsmodule.HTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return analyticsmodule.HTTPAccess{}, analyticsmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.AnalyticsAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues, TenantSlugValues: request.TenantSlugValues,
		Permission: string(request.Permission), Operation: request.Operation,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrAnalyticsAccessAuthentication):
			return analyticsmodule.HTTPAccess{}, analyticsmodule.ErrHTTPAuthentication
		case errors.Is(err, iammodule.ErrAnalyticsAccessForbidden):
			return analyticsmodule.HTTPAccess{}, analyticsmodule.ErrHTTPForbidden
		default:
			return analyticsmodule.HTTPAccess{}, analyticsmodule.ErrHTTPUnavailable
		}
	}
	access, err := analyticsmodule.NewHTTPAccess(decision.PrincipalID(), decision.TenantID(),
		decision.TokenVersion(), decision.AuthorizationVersion(), decision.AuthorizationClock(),
		decision.StepUpAuthenticatedAt(), request.Permission, request.Operation)
	if err != nil {
		return analyticsmodule.HTTPAccess{}, analyticsmodule.ErrHTTPUnavailable
	}
	return access, nil
}

var _ analyticsmodule.HTTPAccessVerifier = (*analyticsAccessVerifier)(nil)
