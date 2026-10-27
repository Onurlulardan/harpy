package api

import (
	"context"
	"errors"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	usagemodule "github.com/Onurlulardan/harpy/core/internal/modules/usage"
)

type usageIAMAccessFixture struct {
	decision iammodule.UsageAccessDecision
	err      error
	request  iammodule.UsageAccessRequest
}

func (fixture *usageIAMAccessFixture) Verify(
	_ context.Context, request iammodule.UsageAccessRequest,
) (iammodule.UsageAccessDecision, error) {
	fixture.request = request
	return fixture.decision, fixture.err
}

func TestUsageAccessVerifierPreservesIAMDecision(t *testing.T) {
	actor := [16]byte{0x10, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	tenant := [16]byte{0x20, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	environment := [16]byte{0x30, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	decision, err := iammodule.NewUsageAccessDecision(actor, tenant, tenant, environment,
		3, 5, 7, "tenant", "onpremise", "onpremise_primary", false,
		"analytics.usage.read", "usage.tenant.overview")
	if err != nil {
		t.Fatal(err)
	}
	iam := &usageIAMAccessFixture{decision: decision}
	verifier, err := newUsageAccessVerifier(iam)
	if err != nil {
		t.Fatal(err)
	}
	request := usagemodule.HTTPAccessRequest{Scope: "tenant",
		Permission: "analytics.usage.read", Operation: "usage.tenant.overview"}
	access, err := verifier.Verify(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if access.ActorID() != actor || access.SubjectTenantID() != tenant ||
		access.EnvironmentID() != environment || access.TokenVersion() != 3 ||
		access.AuthorizationVersion() != 5 || access.AuthorityClock() != 7 ||
		!access.Allows(request.Scope, request.Operation) {
		t.Fatalf("Usage access decision drifted: %#v", access)
	}
	if iam.request.Scope != "tenant" || iam.request.Permission != request.Permission ||
		iam.request.Operation != request.Operation {
		t.Fatalf("Usage IAM request drifted: %#v", iam.request)
	}
}

func TestUsageAccessVerifierMapsIAMFailuresFailClosed(t *testing.T) {
	tests := []struct {
		name     string
		in, want error
	}{
		{"authentication", iammodule.ErrUsageAccessAuthentication, usagemodule.ErrHTTPAuthentication},
		{"forbidden", iammodule.ErrUsageAccessForbidden, usagemodule.ErrHTTPForbidden},
		{"unavailable", iammodule.ErrUsageAccessUnavailable, usagemodule.ErrHTTPUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			verifier, err := newUsageAccessVerifier(&usageIAMAccessFixture{err: test.in})
			if err != nil {
				t.Fatal(err)
			}
			_, err = verifier.Verify(context.Background(), usagemodule.HTTPAccessRequest{})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
