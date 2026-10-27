package actorassertionv1

import (
	"bytes"
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/google/uuid"
)

type assertionFixture struct {
	private     ed25519.PrivateKey
	public      ed25519.PublicKey
	decision    DecisionEvidence
	expectation Expectation
	now         time.Time
}

func newAssertionFixture(t *testing.T) assertionFixture {
	t.Helper()
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, ed25519.SeedSize))
	public := append(ed25519.PublicKey(nil), private[ed25519.SeedSize:]...)
	target, err := TargetBindingSHA256("POST", "/api/forms/record", "record-17")
	if err != nil {
		t.Fatal(err)
	}
	decision := DecisionEvidence{
		TenantID:     uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		SubjectID:    uuid.MustParse("20000000-0000-4000-8000-000000000002"),
		PermissionID: uuid.MustParse("30000000-0000-4000-8000-000000000003"),
		TargetDomain: DomainDocuments, ResourceKey: "form.record", ActionKey: "create",
		TargetBindingSHA256: target, IAMGroupGeneration: 11, IAMGeneration: 12,
		TargetGeneration: 13, TokenVersion: 7, AuthorizationVersion: 19,
		AuthorizationClock: 23,
	}
	expectation := Expectation{TenantID: decision.TenantID, TargetDomain: decision.TargetDomain,
		ResourceKey: decision.ResourceKey, ActionKey: decision.ActionKey,
		TargetBindingSHA256: decision.TargetBindingSHA256,
		IAMGroupGeneration:  decision.IAMGroupGeneration, IAMGeneration: decision.IAMGeneration,
		TargetGeneration: decision.TargetGeneration}
	return assertionFixture{private: private, public: public, decision: decision,
		expectation: expectation, now: time.Unix(2_000_000_000, 0).UTC()}
}

func (fixture assertionFixture) issue(t *testing.T) (Credential, Claims) {
	t.Helper()
	signer, err := NewSigner("edge-2026-07", fixture.private, bytes.NewReader(bytes.Repeat([]byte{5}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	defer signer.Close()
	credential, err := signer.Issue(fixture.decision, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := credential.Wire()
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewVerifier(map[string]ed25519.PublicKey{"edge-2026-07": fixture.public})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := verifier.Verify(wire, fixture.expectation, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	return credential, claims
}
