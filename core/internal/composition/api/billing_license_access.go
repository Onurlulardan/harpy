package api

import (
	"context"
	"errors"
	"fmt"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type billingLicenseIAMAccess interface {
	Verify(context.Context, iammodule.PlatformAccessRequest) (
		iammodule.PlatformAccessDecision,
		error,
	)
}

type billingLicenseAccessVerifier struct{ iam billingLicenseIAMAccess }

func newBillingLicenseAccessVerifier(
	iam billingLicenseIAMAccess,
) (*billingLicenseAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("billing license IAM access runtime is required")
	}
	return &billingLicenseAccessVerifier{iam: iam}, nil
}

func (verifier *billingLicenseAccessVerifier) Verify(
	ctx context.Context,
	request billingmodule.PlatformLicenseHTTPAccessRequest,
) (billingmodule.PlatformLicenseHTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return billingmodule.PlatformLicenseHTTPAccess{},
			billingmodule.ErrPlatformLicenseHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.PlatformAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues,
		TenantSlugValues:    request.TenantSlugValues,
		Capability:          request.Capability, ExpectedMode: request.ExpectedMode,
		RequireStepUp: false,
	})
	if err != nil {
		return billingmodule.PlatformLicenseHTTPAccess{},
			mapBillingLicenseAccessError(err)
	}
	return billingmodule.PlatformLicenseHTTPAccess{
		ActorID: decision.SubjectID.String(), TenantID: decision.TenantID.String(),
		SessionID: decision.SessionID.String(), TokenVersion: decision.TokenVersion,
		AuthorizationVersion: decision.AuthorizationVersion,
		PlatformClock:        decision.PlatformClock,
	}, nil
}

func mapBillingLicenseAccessError(err error) error {
	switch {
	case errors.Is(err, iammodule.ErrPlatformAccessAuthentication):
		return billingmodule.ErrPlatformLicenseHTTPAuthentication
	case errors.Is(err, iammodule.ErrPlatformAccessForbidden):
		return billingmodule.ErrPlatformLicenseHTTPForbidden
	default:
		return billingmodule.ErrPlatformLicenseHTTPUnavailable
	}
}

var _ billingmodule.PlatformLicenseHTTPAccessVerifier = (*billingLicenseAccessVerifier)(nil)
