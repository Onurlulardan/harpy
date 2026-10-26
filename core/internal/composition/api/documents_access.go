package api

import (
	"context"
	"errors"
	"fmt"

	documentsmodule "github.com/Onurlulardan/harpy/core/internal/modules/documents"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type documentsIAMAccess interface {
	Verify(context.Context, iammodule.DocumentsAccessRequest) (
		iammodule.DocumentsAccessDecision, error,
	)
}

type documentsAccessVerifier struct{ iam documentsIAMAccess }

func newDocumentsAccessVerifier(iam documentsIAMAccess) (*documentsAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("Documents IAM access is required")
	}
	return &documentsAccessVerifier{iam: iam}, nil
}

func (verifier *documentsAccessVerifier) Verify(ctx context.Context,
	request documentsmodule.HTTPAccessRequest,
) (documentsmodule.HTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return documentsmodule.HTTPAccess{}, documentsmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.DocumentsAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues, TenantSlugValues: request.TenantSlugValues,
		Permission: request.Permission, Operation: request.Operation,
		PurposeTarget: request.PurposeTarget,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrDocumentsAccessAuthentication):
			return documentsmodule.HTTPAccess{}, documentsmodule.ErrHTTPAuthentication
		case errors.Is(err, iammodule.ErrDocumentsAccessForbidden):
			return documentsmodule.HTTPAccess{}, documentsmodule.ErrHTTPForbidden
		default:
			return documentsmodule.HTTPAccess{}, documentsmodule.ErrHTTPUnavailable
		}
	}
	access, err := documentsmodule.NewHTTPAccess(decision.PrincipalID(), decision.TenantID(),
		decision.TokenVersion(), decision.AuthorizationVersion(), decision.AuthorizationClock(),
		decision.StepUpAuthenticatedAt(), request.Permission, request.Operation)
	if err != nil {
		return documentsmodule.HTTPAccess{}, documentsmodule.ErrHTTPUnavailable
	}
	return access, nil
}

var _ documentsmodule.HTTPAccessVerifier = (*documentsAccessVerifier)(nil)
