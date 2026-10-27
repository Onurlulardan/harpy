package api

import (
	"context"
	"errors"
	"fmt"

	catalogmodule "github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type catalogIAMAccess interface {
	Verify(context.Context, iammodule.CatalogAccessRequest) (
		iammodule.CatalogAccessDecision, error,
	)
}

type catalogAccessVerifier struct{ iam catalogIAMAccess }

func newCatalogAccessVerifier(iam catalogIAMAccess) (*catalogAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("Catalog IAM access is required")
	}
	return &catalogAccessVerifier{iam: iam}, nil
}

func (verifier *catalogAccessVerifier) Verify(
	ctx context.Context, request catalogmodule.HTTPAccessRequest,
) (catalogmodule.HTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return catalogmodule.HTTPAccess{}, catalogmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.CatalogAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues,
		TenantSlugValues:    request.TenantSlugValues,
		Permission:          string(request.Permission),
		Operation:           request.Operation,
		PurposeTarget:       request.PurposeTarget,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrCatalogAccessAuthentication):
			return catalogmodule.HTTPAccess{}, catalogmodule.ErrHTTPAuthentication
		case errors.Is(err, iammodule.ErrCatalogAccessForbidden):
			return catalogmodule.HTTPAccess{}, catalogmodule.ErrHTTPForbidden
		default:
			return catalogmodule.HTTPAccess{}, catalogmodule.ErrHTTPUnavailable
		}
	}
	access, err := catalogmodule.NewHTTPAccessWithAuthorizationEvidence(
		decision.PrincipalID(), decision.TenantID(),
		decision.TokenVersion(), decision.AuthorizationVersion(),
		decision.AuthorizationClock(), decision.StepUpAuthenticatedAt(),
		decision.PermissionID(), decision.AuthorizationEvidenceSHA256(),
		decision.StepUpPurposeSHA256(), request.Permission, request.Operation,
	)
	if err != nil {
		return catalogmodule.HTTPAccess{}, catalogmodule.ErrHTTPUnavailable
	}
	return access, nil
}

var _ catalogmodule.HTTPAccessVerifier = (*catalogAccessVerifier)(nil)
