package serviceprincipal

import (
	"testing"
	"time"
)

func TestVerificationEvidenceMatchesExactPinnedRequest(t *testing.T) {
	request := verificationFixture()
	evidence := VerificationEvidence{
		PrincipalID: request.PrincipalID, PrincipalRevision: 3,
		AuthorizationVersion: 8, GrantID: request.GrantID,
		GrantRevision: request.GrantRevision, Capability: request.Capability,
		Scope: request.Scope, ManifestSHA256: request.ManifestSHA256,
		VerifiedAt: request.VerifyAt,
	}
	if !evidence.Matches(request) {
		t.Fatal("exact evidence did not match")
	}
	evidence.AuthorizationVersion = 0
	if evidence.Matches(request) {
		t.Fatal("revision-less evidence matched")
	}
}

func TestVerificationRequestRejectsUnstableIdentityAndClockDrift(t *testing.T) {
	request := verificationFixture()
	request.PrincipalID[6] = 0x40
	if request.Validate() == nil {
		t.Fatal("non-UUIDv7 principal was accepted")
	}
	request = verificationFixture()
	request.VerifyAt = request.VerifyAt.Add(time.Nanosecond)
	if request.Validate() == nil {
		t.Fatal("sub-microsecond verification time was accepted")
	}
}

func verificationFixture() VerificationRequest {
	principal := [16]byte{0x01, 0x9b, 0, 0, 0, 0, 0x70, 1, 0x80, 0, 0, 0, 0, 0, 0, 1}
	grant := principal
	grant[15] = 2
	resource := principal
	resource[15] = 3
	return VerificationRequest{
		PrincipalID: principal, GrantID: grant, GrantRevision: 4,
		Capability: "workflow.instance.execute",
		Scope: Scope{
			Kind: "resource", Owner: "workflow", ResourceKind: "definition",
			ResourceID: resource,
		},
		ManifestSHA256: [32]byte{1},
		VerifyAt:       time.Date(2026, 7, 18, 12, 0, 0, 123000, time.UTC),
	}
}
