package api

import (
	"context"
	"errors"
	"testing"

	auditmodule "github.com/Onurlulardan/harpy/core/internal/modules/audit"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
)

type auditIAMAccessFixture struct {
	decision iammodule.AuditAccessDecision
	err      error
	request  iammodule.AuditAccessRequest
}

func (fixture *auditIAMAccessFixture) Verify(
	_ context.Context, request iammodule.AuditAccessRequest,
) (iammodule.AuditAccessDecision, error) {
	fixture.request = request
	return fixture.decision, fixture.err
}

func TestAuditAccessVerifierPreservesIAMDecision(t *testing.T) {
	actor := [16]byte{0x10, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	tenant := [16]byte{0x20, 1, 2, 3, 4, 5, 0x40, 7, 0x80, 9, 10, 11, 12, 13, 14, 15}
	decision, err := iammodule.NewAuditAccessDecision(
		actor, tenant, 3, 5, 7, "changelog.read", "changelog.list",
	)
	if err != nil {
		t.Fatal(err)
	}
	iam := &auditIAMAccessFixture{decision: decision}
	verifier, err := newAuditAccessVerifier(iam)
	if err != nil {
		t.Fatal(err)
	}
	request := auditmodule.HTTPAccessRequest{
		Permission: auditmodule.PermissionRead, Operation: "changelog.list",
	}
	access, err := verifier.Verify(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if access.PrincipalID() != actor || access.TenantID() != tenant ||
		access.TokenVersion() != 3 || access.AuthorizationVersion() != 5 ||
		access.AuthorizationClock() != 7 ||
		!access.ValidFor(request.Permission, request.Operation) {
		t.Fatalf("Audit access decision drifted: %#v", access)
	}
	if iam.request.Permission != "changelog.read" || iam.request.Operation != request.Operation {
		t.Fatalf("Audit IAM request drifted: %#v", iam.request)
	}
}

func TestAuditAccessVerifierMapsIAMFailuresFailClosed(t *testing.T) {
	tests := []struct {
		name     string
		in, want error
	}{
		{"authentication", iammodule.ErrAuditAccessAuthentication, auditmodule.ErrHTTPAuthentication},
		{"forbidden", iammodule.ErrAuditAccessForbidden, auditmodule.ErrHTTPForbidden},
		{"unavailable", iammodule.ErrAuditAccessUnavailable, auditmodule.ErrHTTPUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			verifier, err := newAuditAccessVerifier(&auditIAMAccessFixture{err: test.in})
			if err != nil {
				t.Fatal(err)
			}
			_, err = verifier.Verify(context.Background(), auditmodule.HTTPAccessRequest{})
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
