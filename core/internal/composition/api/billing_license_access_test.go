package api

import (
	"context"
	"testing"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type billingLicenseIAMStub struct {
	request iammodule.PlatformAccessRequest
}

func (stub *billingLicenseIAMStub) Verify(
	_ context.Context,
	request iammodule.PlatformAccessRequest,
) (iammodule.PlatformAccessDecision, error) {
	stub.request = request
	return iammodule.PlatformAccessDecision{
		DeploymentMode: request.ExpectedMode, Capability: request.Capability,
		TokenVersion: 1, AuthorizationVersion: 2, PlatformClock: 3,
	}, nil
}

func TestBillingLicenseAccessNeverRequestsStepUp(t *testing.T) {
	iam := &billingLicenseIAMStub{}
	verifier, err := newBillingLicenseAccessVerifier(iam)
	if err != nil {
		t.Fatal(err)
	}
	access, err := verifier.Verify(
		t.Context(),
		billingmodule.PlatformLicenseHTTPAccessRequest{
			AuthorizationValues: []string{"Bearer token"},
			TenantIDValues: []string{
				"00000000-0000-4000-8000-000000000002",
			},
			Capability: "platform.license.refresh", ExpectedMode: "onpremise",
		},
	)
	if err != nil {
		t.Fatalf("access=%+v err=%v", access, err)
	}
	if iam.request.RequireStepUp ||
		iam.request.Capability != "platform.license.refresh" ||
		iam.request.ExpectedMode != "onpremise" {
		t.Fatalf("IAM request=%+v", iam.request)
	}
}
