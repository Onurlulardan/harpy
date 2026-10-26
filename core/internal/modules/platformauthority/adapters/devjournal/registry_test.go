package devjournal

import (
	"context"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

const (
	journalEnvironment = "11111111-1111-4111-8111-111111111111"
	journalBindingID   = "22222222-2222-4222-8222-222222222222"
	journalOperation   = "33333333-3333-4333-8333-333333333333"
	journalOwner       = "44444444-4444-4444-8444-444444444444"
)

type fixedClock struct{ value time.Time }

func (clock fixedClock) Now() time.Time { return clock.value }

type keyResolver map[string]ed25519.PublicKey

func (resolver keyResolver) ResolveProposalKey(
	_ context.Context, kid string,
) (ed25519.PublicKey, error) {
	if key, found := resolver[kid]; found {
		return key, nil
	}
	return nil, os.ErrNotExist
}

type journalFixture struct {
	registry        *Registry
	proposalPrivate ed25519.PrivateKey
	proof           artifact.VerifiedBindingProof
	state           artifact.CanonicalState
	now             time.Time
	path            string
	recovery        recoveryv1.Bundle
	resolver        keyResolver
}

func newJournalFixture(t *testing.T) journalFixture {
	t.Helper()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	proposalPublic, proposalPrivate := journalKey(1)
	_, registryPrivate := journalKey(2)
	recovery, err := recoveryv1.Build(journalEnvironment, now.Add(-time.Hour),
		journalRecoveryMaterial(proposalPublic, 3))
	if err != nil {
		t.Fatal(err)
	}
	proposalKID, _ := recovery.KeyID(recoveryv1.PurposeRecoveryControl)
	resolver := keyResolver{proposalKID: proposalPublic}
	path := filepath.Join(t.TempDir(), "authority.journal")
	registry, err := New(
		path, "registry.dev.v1", registryPrivate,
		resolver, fixedClock{now},
	)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := artifact.SignBindingProof(artifact.BindingProof{
		Schema: artifact.BindingSchema, EnvironmentID: journalEnvironment,
		LicenseBindingID:          journalBindingID,
		FinalizationReceiptSHA256: strings.Repeat("a", 64),
		RecoveryTrustBundleSHA256: recovery.SHA256(), ControlKID: proposalKID,
		RegistryNonce: "55555555-5555-4555-8555-555555555555",
		OperationID:   "66666666-6666-4666-8666-666666666666",
	}, proposalPrivate)
	if err != nil {
		t.Fatal(err)
	}
	owner := journalOwner
	ownerGeneration := int64(1)
	adminDigest := strings.Repeat("c", 64)
	state, err := artifact.CanonicalizeState(artifact.State{
		Schema: artifact.StateSchema, EnvironmentID: journalEnvironment,
		DeploymentMode: artifact.ModeOnPremise, RegistryBindingID: journalBindingID,
		ScopeKind: artifact.ScopeEnvironment, ScopeID: journalEnvironment,
		ProposalSignerGeneration: 1, ProposalSignerKID: proposalKID,
		OwnerBindingID: &owner, OwnerBindingGeneration: &ownerGeneration,
		PrimaryAdminBindingSetSHA256: &adminDigest,
		RecoveryTrustBundleSHA256:    stringPointer(recovery.SHA256()),
	})
	if err != nil {
		t.Fatal(err)
	}
	return journalFixture{
		registry: registry, proposalPrivate: proposalPrivate, proof: proof,
		state: state, now: now, path: path, recovery: recovery, resolver: resolver,
	}
}

func TestJournalPinsBindingAndCASesReplay(t *testing.T) {
	fixture := newJournalFixture(t)
	ctx := context.Background()
	if err := fixture.registry.EnsureBinding(ctx, fixture.proof); err != nil {
		t.Fatal(err)
	}
	if err := fixture.registry.EnsureBinding(ctx, fixture.proof); err != nil {
		t.Fatalf("idempotent binding failed: %v", err)
	}
	proposal := fixture.signProposal(t, journalOperation, 1, nil)
	first, err := fixture.registry.ExchangeProposal(ctx, proposal)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := fixture.registry.ExchangeProposal(ctx, proposal)
	if err != nil || string(first) != string(replayed) {
		t.Fatalf("operation replay changed ack: %v", err)
	}
	head, err := fixture.registry.LatestHead(ctx, ports.HeadKey{
		RegistryBindingID: journalBindingID,
		ScopeKind:         artifact.ScopeEnvironment, ScopeID: journalEnvironment,
	})
	if err != nil || head.Generation != 1 {
		t.Fatalf("head = %+v, err = %v", head, err)
	}
	if _, err := fixture.registry.ExchangeProposal(
		ctx, fixture.signProposal(t, "77777777-7777-4777-8777-777777777777", 1, nil),
	); err == nil {
		t.Fatal("ABA generation replay was accepted")
	}
}

func TestJournalRejectsRebindAndTamper(t *testing.T) {
	fixture := newJournalFixture(t)
	ctx := context.Background()
	if err := fixture.registry.EnsureBinding(ctx, fixture.proof); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.registry.ExchangeProposal(
		ctx, fixture.signProposal(t, journalOperation, 1, nil),
	); err != nil {
		t.Fatal(err)
	}
	document, err := os.ReadFile(fixture.path)
	if err != nil {
		t.Fatal(err)
	}
	document[len(document)/2] ^= 1
	if err := os.WriteFile(fixture.path, document, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.registry.LatestHead(ctx, ports.HeadKey{
		RegistryBindingID: journalBindingID,
		ScopeKind:         artifact.ScopeEnvironment, ScopeID: journalEnvironment,
	}); err == nil {
		t.Fatal("tampered journal was accepted")
	}
}

func (fixture journalFixture) signProposal(
	t *testing.T, operationID string, generation int64, previous *string,
) artifact.SignedProposal {
	t.Helper()
	proposal, err := artifact.SignProposal(artifact.Proposal{
		Schema: artifact.ProposalSchema, Purpose: "initial_environment_authority",
		OperationID: operationID, EnvironmentID: journalEnvironment,
		DeploymentMode: artifact.ModeOnPremise, ScopeKind: artifact.ScopeEnvironment,
		ScopeID: journalEnvironment, RegistryBindingID: journalBindingID,
		Generation: generation, PreviousHeadSHA256: previous, StateSHA256: fixture.state.SHA256,
		RecoveryTrustBundleSHA256: fixture.state.Value.RecoveryTrustBundleSHA256,
		IssuedAt:                  fixture.now.Format(time.RFC3339),
		ExpiresAt:                 fixture.now.Add(10 * time.Minute).Format(time.RFC3339),
		SignerKID:                 fixture.proof.Value.ControlKID,
	}, fixture.state, artifact.ExpiryOnline, fixture.proposalPrivate)
	if err != nil {
		t.Fatal(err)
	}
	return proposal
}
