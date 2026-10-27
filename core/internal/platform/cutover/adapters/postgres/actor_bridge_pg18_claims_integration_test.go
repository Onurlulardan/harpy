//go:build integration

package postgres

import (
	"crypto/ed25519"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	"github.com/google/uuid"
)

type actorBridgeIssuedCredential struct {
	claims      actorassertionv1.Claims
	wire        string
	expectation actorassertionv1.Expectation
	publicKeys  map[string]ed25519.PublicKey
	currentTime time.Time
}

func issueActorBridgeClaims(
	t *testing.T,
	timepoint time.Time,
	tenantID string,
	authority actorbridge.Authority,
) actorassertionv1.Claims {
	t.Helper()
	return issueActorBridgeCredential(t, timepoint, tenantID, authority).claims
}

func issueActorBridgeCredential(
	t *testing.T,
	timepoint time.Time,
	tenantID string,
	authority actorbridge.Authority,
) actorBridgeIssuedCredential {
	t.Helper()
	targetDigest, err := actorassertionv1.TargetBindingSHA256(
		"POST", "/api/v1/forms/records", "record-17",
	)
	if err != nil {
		t.Fatal(err)
	}
	decision := actorassertionv1.DecisionEvidence{
		TenantID:             uuid.MustParse(tenantID),
		SubjectID:            uuid.MustParse(actorBridgeUserID),
		PermissionID:         uuid.MustParse(actorBridgePermissionID),
		TargetDomain:         actorassertionv1.DomainForms,
		ResourceKey:          "forms.records",
		ActionKey:            "read",
		TargetBindingSHA256:  targetDigest,
		IAMGroupGeneration:   authority.IAMGroupGeneration(),
		IAMGeneration:        authority.IAMGeneration(),
		TargetGeneration:     authority.TargetGeneration(),
		TokenVersion:         1,
		AuthorizationVersion: 1,
		AuthorizationClock:   0,
	}
	return signActorBridgeCredential(t, timepoint, decision)
}

func signActorBridgeDecision(
	t *testing.T, timepoint time.Time, decision actorassertionv1.DecisionEvidence,
) actorassertionv1.Claims {
	t.Helper()
	return signActorBridgeCredential(t, timepoint, decision).claims
}

func signActorBridgeCredential(
	t *testing.T,
	timepoint time.Time,
	decision actorassertionv1.DecisionEvidence,
) actorBridgeIssuedCredential {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = byte(index + 1)
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	clear(seed)
	publicKey := append(ed25519.PublicKey(nil), privateKey[ed25519.SeedSize:]...)
	signer, err := actorassertionv1.NewProductionSigner("edge-2026-07", privateKey)
	clear(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	defer signer.Close()
	credential, err := signer.Issue(decision, timepoint.UTC().Truncate(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	wire, err := credential.Wire()
	if err != nil {
		t.Fatal(err)
	}
	publicKeys := map[string]ed25519.PublicKey{"edge-2026-07": publicKey}
	verifier, err := actorassertionv1.NewVerifier(publicKeys)
	if err != nil {
		t.Fatal(err)
	}
	expectation := actorassertionv1.Expectation{
		TenantID:            decision.TenantID,
		TargetDomain:        decision.TargetDomain,
		ResourceKey:         decision.ResourceKey,
		ActionKey:           decision.ActionKey,
		TargetBindingSHA256: decision.TargetBindingSHA256,
		IAMGroupGeneration:  decision.IAMGroupGeneration,
		IAMGeneration:       decision.IAMGeneration,
		TargetGeneration:    decision.TargetGeneration,
	}
	claims, err := verifier.Verify(wire, expectation, timepoint)
	if err != nil {
		t.Fatal(err)
	}
	return actorBridgeIssuedCredential{claims: claims, wire: wire,
		expectation: expectation, publicKeys: publicKeys, currentTime: timepoint}
}
