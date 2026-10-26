package actorassertionv1

import (
	"crypto/ed25519"
	"errors"
	"testing"
)

func TestVerifierRejectsWrongRequestAndAuthorityBinding(t *testing.T) {
	fixture := newAssertionFixture(t)
	credential, _ := fixture.issue(t)
	wire, _ := credential.Wire()
	verifier, err := NewVerifier(map[string]ed25519.PublicKey{"edge-2026-07": fixture.public})
	if err != nil {
		t.Fatal(err)
	}
	mutations := []Expectation{fixture.expectation, fixture.expectation, fixture.expectation,
		fixture.expectation, fixture.expectation, fixture.expectation}
	mutations[0].TenantID[0] ^= 1
	mutations[1].TargetDomain = DomainWorkflow
	mutations[2].ResourceKey = "workflow.instance"
	mutations[3].ActionKey = "update"
	mutations[4].TargetBindingSHA256[0] ^= 1
	mutations[5].TargetGeneration++
	for index, expectation := range mutations {
		if _, verifyErr := verifier.Verify(wire, expectation, fixture.now); verifyErr == nil {
			t.Fatalf("expectation mutation %d was accepted", index)
		}
	}
}

func TestVerifierEnforcesThirtySecondLifetime(t *testing.T) {
	fixture := newAssertionFixture(t)
	credential, _ := fixture.issue(t)
	wire, _ := credential.Wire()
	verifier, _ := NewVerifier(map[string]ed25519.PublicKey{"edge-2026-07": fixture.public})
	if _, err := verifier.Verify(wire, fixture.expectation, fixture.now.Add(29_999_999_999)); err != nil {
		t.Fatalf("live assertion rejected: %v", err)
	}
	if _, err := verifier.Verify(wire, fixture.expectation, fixture.now.Add(30_000_000_000)); !errors.Is(err, ErrExpired) {
		t.Fatalf("expiry error = %v", err)
	}
	if _, err := verifier.Verify(wire, fixture.expectation, fixture.now.Add(-3_000_000_000)); err == nil {
		t.Fatal("assertion beyond future clock skew was accepted")
	}
}
