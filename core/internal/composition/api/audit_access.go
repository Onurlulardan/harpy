package api

import (
	"context"
	"errors"
	"fmt"

	auditmodule "github.com/Onurlulardan/harpy/core/internal/modules/audit"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type auditIAMAccess interface {
	Verify(context.Context, iammodule.AuditAccessRequest) (
		iammodule.AuditAccessDecision, error,
	)
}

type auditAccessVerifier struct{ iam auditIAMAccess }

func newAuditAccessVerifier(iam auditIAMAccess) (*auditAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("Audit IAM access is required")
	}
	return &auditAccessVerifier{iam: iam}, nil
}

func (verifier *auditAccessVerifier) Verify(
	ctx context.Context, request auditmodule.HTTPAccessRequest,
) (auditmodule.HTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return auditmodule.HTTPAccess{}, auditmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.AuditAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues, TenantSlugValues: request.TenantSlugValues,
		Permission: string(request.Permission), Operation: request.Operation,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrAuditAccessAuthentication):
			return auditmodule.HTTPAccess{}, auditmodule.ErrHTTPAuthentication
		case errors.Is(err, iammodule.ErrAuditAccessForbidden):
			return auditmodule.HTTPAccess{}, auditmodule.ErrHTTPForbidden
		default:
			return auditmodule.HTTPAccess{}, auditmodule.ErrHTTPUnavailable
		}
	}
	access, err := auditmodule.NewHTTPAccess(
		decision.PrincipalID(), decision.TenantID(), decision.TokenVersion(),
		decision.AuthorizationVersion(), decision.AuthorizationClock(),
		request.Permission, request.Operation,
	)
	if err != nil {
		return auditmodule.HTTPAccess{}, auditmodule.ErrHTTPUnavailable
	}
	return access, nil
}

var _ auditmodule.HTTPAccessVerifier = (*auditAccessVerifier)(nil)
