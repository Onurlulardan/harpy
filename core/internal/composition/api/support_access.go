package api

import (
	"context"
	"errors"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	supportmodule "github.com/Onurlulardan/harpy/core/internal/modules/support"
)

type supportIAMAccess interface {
	Verify(context.Context, iammodule.SupportAccessRequest) (
		iammodule.SupportAccessDecision, error,
	)
}

type supportAccessVerifier struct{ iam supportIAMAccess }

func newSupportAccessVerifier(iam supportIAMAccess) (*supportAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("Support IAM access is required")
	}
	return &supportAccessVerifier{iam: iam}, nil
}

func (verifier *supportAccessVerifier) Verify(
	ctx context.Context, request supportmodule.HTTPAccessRequest,
) (supportmodule.HTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return supportmodule.HTTPAccess{}, supportmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.SupportAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues, TenantSlugValues: request.TenantSlugValues,
		Scope: string(request.Scope), Permission: request.Permission, Operation: request.Operation,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrSupportAccessAuthentication):
			return supportmodule.HTTPAccess{}, supportmodule.ErrHTTPAuthentication
		case errors.Is(err, iammodule.ErrSupportAccessForbidden):
			return supportmodule.HTTPAccess{}, supportmodule.ErrHTTPForbidden
		default:
			return supportmodule.HTTPAccess{}, supportmodule.ErrHTTPUnavailable
		}
	}
	access, err := supportmodule.NewHTTPAccess(
		decision.ActorID(), decision.TenantID(), decision.TokenVersion(),
		decision.AuthorizationVersion(), decision.AuthorityClock(),
		supportmodule.HTTPAccessScope(decision.Scope()), request.Permission,
		request.Operation, decision.ReporterName(), decision.ReporterEmail(),
		decision.TenantSlug(), decision.TenantName(),
	)
	if err != nil {
		return supportmodule.HTTPAccess{}, supportmodule.ErrHTTPUnavailable
	}
	return access, nil
}

var _ supportmodule.HTTPAccessVerifier = (*supportAccessVerifier)(nil)
