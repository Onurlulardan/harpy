package api

import (
	"context"
	"errors"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	supportmodule "github.com/Onurlulardan/harpy/core/internal/modules/support"
)

type supportIAMAccessFixture struct {
	decision iammodule.SupportAccessDecision
	err      error
	request  iammodule.SupportAccessRequest
}

func (fixture *supportIAMAccessFixture) Verify(
	_ context.Context, request iammodule.SupportAccessRequest,
) (iammodule.SupportAccessDecision, error) {
	fixture.request = request
	return fixture.decision, fixture.err
}

func TestSupportAccessVerifierPreservesIAMDecision(t *testing.T) {
	actor := [16]byte{0x10, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	tenant := [16]byte{0x20, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	decision, err := iammodule.NewSupportAccessDecision(actor, tenant, 3, 5, 7,
		"reporter", "bug_reports.create", "support.bug.create",
		"Reporter", "reporter@example.com", "tenant", "Tenant")
	if err != nil {
		t.Fatal(err)
	}
	iam := &supportIAMAccessFixture{decision: decision}
	verifier, err := newSupportAccessVerifier(iam)
	if err != nil {
		t.Fatal(err)
	}
	request := supportmodule.HTTPAccessRequest{Scope: supportmodule.HTTPReporterScope,
		Permission: "bug_reports.create", Operation: "support.bug.create"}
	access, err := verifier.Verify(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if access.ActorID() != actor || access.TenantID() != tenant ||
		access.AuthorizationVersion() != 5 || access.AuthorityClock() != 7 ||
		!access.Allows(request.Scope, request.Permission, request.Operation) {
		t.Fatalf("Support access decision drifted: %#v", access)
	}
	if iam.request.Scope != "reporter" || iam.request.Permission != request.Permission ||
		iam.request.Operation != request.Operation {
		t.Fatalf("Support IAM request drifted: %#v", iam.request)
	}
}

func TestSupportAccessVerifierMapsIAMFailuresFailClosed(t *testing.T) {
	tests := []struct {
		name     string
		in, want error
	}{
		{"authentication", iammodule.ErrSupportAccessAuthentication, supportmodule.ErrHTTPAuthentication},
		{"forbidden", iammodule.ErrSupportAccessForbidden, supportmodule.ErrHTTPForbidden},
		{"unavailable", iammodule.ErrSupportAccessUnavailable, supportmodule.ErrHTTPUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			verifier, err := newSupportAccessVerifier(&supportIAMAccessFixture{err: test.in})
			if err != nil {
				t.Fatal(err)
			}
			_, err = verifier.Verify(context.Background(), supportmodule.HTTPAccessRequest{})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
