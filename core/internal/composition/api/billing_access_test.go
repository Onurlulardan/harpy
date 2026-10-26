package api

import (
	"context"
	"errors"
	"testing"

	billingmodule "github.com/Onurlulardan/harpy/core/internal/modules/billing"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type billingIAMAccessStub struct {
	request  iammodule.BillingAccessRequest
	decision iammodule.BillingAccessDecision
	err      error
}

func TestTenantAllDecisionCannotCrossBillingGlobalScopeBoundary(t *testing.T) {
	tenantDecision, err := iammodule.NewTenantBillingAccessDecision(
		[16]byte{1}, [16]byte{2}, 3, 4, 5,
	)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := newBillingAccessVerifier(&billingIAMAccessStub{
		decision: tenantDecision,
	})
	if err != nil {
		t.Fatal(err)
	}
	access, err := verifier.Verify(context.Background(), billingmodule.AccessRequest{
		Operation:  "billing.catalog.offering.list",
		Permission: "platform.billing.read",
	})
	if err == nil || access.Scope() != 0 {
		t.Fatal("tenant-only IAM evidence crossed into a global Billing scope")
	}
}

func (stub *billingIAMAccessStub) Verify(
	_ context.Context,
	request iammodule.BillingAccessRequest,
) (iammodule.BillingAccessDecision, error) {
	stub.request = request
	return stub.decision, stub.err
}

func TestBillingAccessVerifierMapsIAMFailures(t *testing.T) {
	tests := []struct{ source, want error }{
		{iammodule.ErrBillingAccessAuthentication, billingmodule.ErrHTTPAuthentication},
		{iammodule.ErrBillingAccessForbidden, billingmodule.ErrHTTPForbidden},
		{iammodule.ErrBillingAccessUnavailable, billingmodule.ErrHTTPUnavailable},
	}
	for _, test := range tests {
		verifier, _ := newBillingAccessVerifier(&billingIAMAccessStub{err: test.source})
		_, err := verifier.Verify(context.Background(), billingmodule.AccessRequest{
			AuthorizationValues: []string{"Bearer forwarded"},
			Operation:           "billing.subscription.change",
		})
		if !errors.Is(err, test.want) {
			t.Fatalf("mapped error = %v; want %v", err, test.want)
		}
		if verifier.iam.(*billingIAMAccessStub).request.Operation !=
			"billing.subscription.change" {
			t.Fatal("Billing access request was not forwarded")
		}
	}
}
