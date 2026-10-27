package api

import (
	"context"
	"errors"
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type billingIAMAccess interface {
	Verify(
		context.Context,
		iammodule.BillingAccessRequest,
	) (iammodule.BillingAccessDecision, error)
}

type billingAccessVerifier struct{ iam billingIAMAccess }

func newBillingAccessVerifier(iam billingIAMAccess) (*billingAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("Billing IAM access runtime is required")
	}
	return &billingAccessVerifier{iam: iam}, nil
}

func (verifier *billingAccessVerifier) Verify(
	ctx context.Context,
	request billingmodule.AccessRequest,
) (billingmodule.VerifiedHTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return billingmodule.VerifiedHTTPAccess{}, billingmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.BillingAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues,
		TenantSlugValues:    request.TenantSlugValues,
		Permission:          string(request.Permission), Operation: request.Operation,
	})
	if err != nil {
		return billingmodule.VerifiedHTTPAccess{}, mapBillingAccessError(err)
	}
	command, commandAccess := decision.GlobalCommandAccess()
	if commandAccess {
		return billingmodule.NewGlobalCommandVerifiedHTTPAccess(
			decision.PrincipalID(), decision.TokenVersion(),
			decision.AuthorizationVersion(), decision.AuthorizationClock(),
			request.Permission, command,
		)
	}
	snapshot, snapshotAccess := decision.GlobalSnapshotCapability()
	if snapshotAccess {
		return billingmodule.NewGlobalSnapshotVerifiedHTTPAccess(
			decision.PrincipalID(), decision.TokenVersion(),
			decision.AuthorizationVersion(), decision.AuthorizationClock(),
			request.Permission, snapshot,
		)
	}
	return billingmodule.NewTenantVerifiedHTTPAccess(
		decision.PrincipalID(), decision.TenantID(), decision.TokenVersion(),
		decision.AuthorizationVersion(), decision.AuthorizationClock(),
		request.Permission,
	)
}

func mapBillingAccessError(err error) error {
	switch {
	case errors.Is(err, iammodule.ErrBillingAccessAuthentication):
		return billingmodule.ErrHTTPAuthentication
	case errors.Is(err, iammodule.ErrBillingAccessForbidden):
		return billingmodule.ErrHTTPForbidden
	default:
		return billingmodule.ErrHTTPUnavailable
	}
}

var _ billingmodule.AccessVerifier = (*billingAccessVerifier)(nil)
