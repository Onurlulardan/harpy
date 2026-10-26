package actorassertionv1

import (
	"fmt"
	"testing"
)

func TestDedicatedEd25519AssertionRoundTrip(t *testing.T) {
	fixture := newAssertionFixture(t)
	credential, claims := fixture.issue(t)
	if claims.Decision != fixture.decision || claims.ExpiresAt-claims.IssuedAt != 30 ||
		claims.NotBefore != claims.IssuedAt || claims.JTI.Version() != 4 {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if fmt.Sprint(credential) != redacted || fmt.Sprintf("%#v", credential) != redacted {
		t.Fatal("credential formatter exposed the assertion")
	}
	if _, err := credential.MarshalJSON(); err == nil {
		t.Fatal("credential JSON serialization was accepted")
	}
}

func TestSignerWipesPrivateKeyAndFailsClosed(t *testing.T) {
	fixture := newAssertionFixture(t)
	signer, err := NewProductionSigner("edge-2026-07", fixture.private)
	if err != nil {
		t.Fatal(err)
	}
	signer.Close()
	if _, err := signer.Issue(fixture.decision, fixture.now); err != ErrClosed {
		t.Fatalf("closed signer error = %v", err)
	}
	if len(signer.key) != 0 {
		t.Fatal("closed signer retained private key bytes")
	}
}

func TestDecisionDigestBindsEveryAuthorizationDimension(t *testing.T) {
	fixture := newAssertionFixture(t)
	want, err := fixture.decision.Digest()
	if err != nil {
		t.Fatal(err)
	}
	mutations := []DecisionEvidence{fixture.decision, fixture.decision, fixture.decision,
		fixture.decision, fixture.decision, fixture.decision}
	mutations[0].TokenVersion++
	mutations[1].AuthorizationVersion++
	mutations[2].AuthorizationClock++
	mutations[3].TargetGeneration++
	mutations[4].ActionKey = "update"
	mutations[5].TargetBindingSHA256[0] ^= 1
	for index, mutation := range mutations {
		got, digestErr := mutation.Digest()
		if digestErr != nil || got == want {
			t.Fatalf("mutation %d was not bound to decision digest", index)
		}
	}
}
