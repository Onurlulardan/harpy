package accesstoken

import (
	"context"
	"strings"
	"testing"
)

func TestAccessTokenRoundTrip(t *testing.T) {
	fixture := newTokenFixture(t)
	signer := fixture.signer(t)
	defer signer.Close()
	token, err := signer.Sign(fixture.claims(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(token) > MaximumWireSize || strings.Contains(token, "=") {
		t.Fatalf("noncanonical compact token: %q", token)
	}
	verified, err := fixture.verifier(t).Verify(context.Background(), token, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	if verified.Subject() != fixture.input.Subject || verified.Tenant() != fixture.input.Tenant ||
		verified.Client() != ClientOperatorConsole || verified.TokenVersion() != 3 {
		t.Fatalf("unexpected verified claims: %#v", verified)
	}
	if _, exists := verified.PortalBinding(); exists {
		t.Fatal("operator token unexpectedly has portal binding")
	}
	segments, _ := splitCompact(token)
	payload, _ := decodeSegment(segments.payload)
	if strings.Contains(string(payload), "role_ids") {
		t.Fatal("access token emitted a role snapshot")
	}
}

func TestOperatorStepUpRoundTrip(t *testing.T) {
	fixture := newTokenFixture(t)
	fixture.input.Assurance = AssuranceStepUp
	fixture.input.Methods = []AuthMethod{AuthMethodPassword, AuthMethodWebAuthn}
	fixture.input.StepUpPurposeHash = make([]byte, 32)
	fixture.input.StepUpPurposeHash[0] = 9
	signer := fixture.signer(t)
	defer signer.Close()
	token, err := signer.Sign(fixture.claims(t))
	if err != nil {
		t.Fatal(err)
	}
	verified, err := fixture.verifier(t).Verify(context.Background(), token, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := verified.PortalBinding(); ok {
		t.Fatal("operator step-up carried a portal binding")
	}
	if hash, ok := verified.StepUpPurposeHash(); !ok || hash[0] != 9 {
		t.Fatal("step-up purpose was not preserved")
	}
}
