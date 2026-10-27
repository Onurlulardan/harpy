package api

import (
	"context"
	"errors"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	workmodule "github.com/Onurlulardan/harpy/core/internal/modules/work"
)

type workIAMAccess interface {
	Verify(context.Context, iammodule.WorkAccessRequest) (
		iammodule.WorkAccessDecision, error,
	)
}

type workAccessVerifier struct{ iam workIAMAccess }

func newWorkAccessVerifier(iam workIAMAccess) (*workAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("Work IAM access is required")
	}
	return &workAccessVerifier{iam: iam}, nil
}

func (verifier *workAccessVerifier) Verify(
	ctx context.Context, request workmodule.HTTPAccessRequest,
) (workmodule.HTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return workmodule.HTTPAccess{}, workmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.WorkAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues,
		TenantSlugValues:    request.TenantSlugValues,
		Authority:           request.Authority, Operation: request.Operation,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrWorkAccessAuthentication):
			return workmodule.HTTPAccess{}, workmodule.ErrHTTPAuthentication
		case errors.Is(err, iammodule.ErrWorkAccessForbidden):
			return workmodule.HTTPAccess{}, workmodule.ErrHTTPForbidden
		default:
			return workmodule.HTTPAccess{}, workmodule.ErrHTTPUnavailable
		}
	}
	access, err := workmodule.NewHTTPAccessWithEvidence(
		decision.PrincipalID(), decision.TenantID(), decision.TokenVersion(),
		decision.AuthorizationVersion(), request.Authority, request.Operation,
		firstHeader(request.ConsumerIDValues), firstHeader(request.BuildIDValues),
		request.Compatibility, decision.ProtectedAdmin(),
	)
	if err != nil {
		return workmodule.HTTPAccess{}, workmodule.ErrHTTPUnavailable
	}
	return access, nil
}

func firstHeader(values []string) string {
	if len(values) != 1 {
		return ""
	}
	return values[0]
}

var _ workmodule.HTTPAccessVerifier = (*workAccessVerifier)(nil)
