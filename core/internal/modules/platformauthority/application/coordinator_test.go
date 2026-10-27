package application

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

var testNow = time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)

const (
	testEnvironmentID = "11111111-1111-4111-8111-111111111111"
	testBindingID     = "22222222-2222-4222-8222-222222222222"
	testOwnerID       = "33333333-3333-4333-8333-333333333333"
	testOperationID   = "44444444-4444-4444-8444-444444444444"
)

type testClock struct{}

func (testClock) Now() time.Time { return testNow }

type coordinatorFixture struct {
	service  *Coordinator
	store    *memoryStore
	registry *fakeRegistry
	state    api.State
	recovery recoveryv1.Bundle
}

type testProposalSigner struct {
	private ed25519.PrivateKey
	public  ed25519.PublicKey
	kid     string
}

func (signer testProposalSigner) KID() string { return signer.kid }
func (signer testProposalSigner) PublicKey(context.Context) (ed25519.PublicKey, error) {
	return append(ed25519.PublicKey(nil), signer.public...), nil
}
func (signer testProposalSigner) Sign(_ context.Context, message []byte) ([]byte, error) {
	return ed25519.Sign(signer.private, message), nil
}

func newCoordinatorFixture(t *testing.T) coordinatorFixture {
	t.Helper()
	releasePublic, releasePrivate := testKey(1)
	proposalPublic, proposalPrivate := testKey(2)
	registryPublic, registryPrivate := testKey(3)
	recovery, err := recoveryv1.Build(testEnvironmentID, testNow.Add(-time.Hour),
		[]recoveryv1.PublicMaterial{
			{Purpose: recoveryv1.PurposeRecoveryManifest, PublicKey: keyArray(testKeyPublic(7))},
			{Purpose: recoveryv1.PurposeRecoveryControl, PublicKey: keyArray(proposalPublic)},
			{Purpose: recoveryv1.PurposeRestoreReceipt, PublicKey: keyArray(testKeyPublic(8))},
		})
	if err != nil {
		t.Fatal(err)
	}
	proposalKID, _ := recovery.KeyID(recoveryv1.PurposeRecoveryControl)
	bundleBytes, err := artifact.SignTrustBundle(artifact.TrustBundle{
		Schema: artifact.TrustBundleSchema, RegistryKind: artifact.RegistryDevelopment,
		TrustGeneration: 1, Active: artifact.TrustKey{
			KID: "registry.v1", Algorithm: artifact.AlgorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(registryPublic),
			NotBefore: testNow.Add(-time.Hour).Format(time.RFC3339),
			SignUntil: testNow.Add(time.Hour).Format(time.RFC3339),
		}, Retired: []artifact.RetiredTrustKey{},
	}, releasePrivate)
	if err != nil {
		t.Fatal(err)
	}
	trust, err := artifact.VerifyTrustBundle(
		bundleBytes, artifact.RegistryDevelopment, testNow, releasePublic,
	)
	if err != nil {
		t.Fatal(err)
	}
	signer := testProposalSigner{private: proposalPrivate, public: proposalPublic, kid: proposalKID}
	store := newMemoryStore()
	recoveryDigest := recovery.SHA256()
	proposalDigest := sha256.Sum256(proposalPublic)
	store.bindings[testBindingID] = ports.BindingRecord{
		BindingID: testBindingID, EnvironmentID: testEnvironmentID,
		DeploymentMode: artifact.ModeOnPremise, RegistryKind: artifact.RegistryDevelopment,
		RecoveryTrustBundleSHA256: &recoveryDigest, SignerGeneration: 1,
		SignerKID: proposalKID, SignerPublicKeySHA256: hex.EncodeToString(proposalDigest[:]),
		SignerStatus: "active",
	}
	registry := &fakeRegistry{
		privateKey: registryPrivate, kid: "registry.v1",
		kind: artifact.RegistryDevelopment, acks: map[string]artifact.VerifiedAck{},
	}
	service, err := NewCoordinator(
		store, registry, signer, testClock{}, trust, artifact.RegistryDevelopment,
	)
	if err != nil {
		t.Fatal(err)
	}
	owner := testOwnerID
	ownerGeneration := int64(1)
	adminDigest := strings.Repeat("a", 64)
	state := api.State{
		Schema: artifact.StateSchema, EnvironmentID: testEnvironmentID,
		DeploymentMode: api.ModeOnPremise, RegistryBindingID: testBindingID,
		ScopeKind: api.ScopeEnvironment, ScopeID: testEnvironmentID,
		ProposalSignerGeneration: 1, ProposalSignerKID: proposalKID,
		OwnerBindingID: &owner, OwnerBindingGeneration: &ownerGeneration,
		PrimaryAdminBindingSetSHA256: &adminDigest,
		RecoveryTrustBundleSHA256:    &recoveryDigest,
	}
	return coordinatorFixture{service: service, store: store, registry: registry,
		state: state, recovery: recovery}
}

func TestInitialAuthorityCrashReplayAndOperationConflict(t *testing.T) {
	fixture := newCoordinatorFixture(t)
	request := api.InitialAuthorityRequest{
		OperationID: testOperationID, State: fixture.state,
		IssuedAt: testNow, Policy: api.ExpiryOnline,
	}
	prepared, err := fixture.service.PrepareInitial(context.Background(), request)
	if err != nil || fixture.registry.exchanges != 0 || prepared.StateSHA256 == "" ||
		prepared.ProposalSHA256 == "" {
		t.Fatalf("prepare = %+v, exchanges=%d, err=%v", prepared, fixture.registry.exchanges, err)
	}
	fixture.store.failCommitOnce = true
	if _, err := fixture.service.AcknowledgePrepared(
		context.Background(), prepared.OperationID,
	); err == nil {
		t.Fatal("commit crash was hidden")
	}
	receipt, err := fixture.service.AcknowledgePrepared(
		context.Background(), prepared.OperationID,
	)
	if err != nil || receipt.Generation != 1 || fixture.registry.exchanges != 1 {
		t.Fatalf("crash replay = %+v, exchanges=%d, err=%v",
			receipt, fixture.registry.exchanges, err)
	}
	replayed, err := fixture.service.AcknowledgePrepared(
		context.Background(), prepared.OperationID,
	)
	if err != nil || replayed.AckSHA256 != receipt.AckSHA256 || fixture.registry.exchanges != 1 {
		t.Fatal("completed operation replay was not byte-stable")
	}
	changed := request
	changed.State.ProposalSignerGeneration = 2
	if _, err := fixture.service.PrepareInitial(
		context.Background(), changed,
	); !errors.Is(err, api.ErrConflict) {
		t.Fatalf("operation ID conflict was accepted: %v", err)
	}
}

func TestNoAckNoHeadAndHigherGenerationCompensation(t *testing.T) {
	fixture := newCoordinatorFixture(t)
	fixture.registry.fail = true
	request := api.InitialAuthorityRequest{
		OperationID: testOperationID, State: fixture.state,
		IssuedAt: testNow, Policy: api.ExpiryOnline,
	}
	prepared, err := fixture.service.PrepareInitial(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.AcknowledgePrepared(
		context.Background(), prepared.OperationID,
	); err == nil {
		t.Fatal("registry outage was accepted")
	}
	if len(fixture.store.heads) != 0 {
		t.Fatal("head advanced without acknowledgement")
	}
	fixture.registry.fail = false
	initial, err := fixture.service.AcknowledgePrepared(
		context.Background(), prepared.OperationID,
	)
	if err != nil {
		t.Fatal(err)
	}
	compensation, err := fixture.service.CompensateUncommitted(
		context.Background(), api.CompensationRequest{
			OperationID: "55555555-5555-4555-8555-555555555555",
			State:       fixture.state, CompensatesAckSHA256: initial.AckSHA256,
			IssuedAt: testNow, Policy: api.ExpiryOnline,
		})
	if err != nil || compensation.Generation != 2 {
		t.Fatalf("compensation = %+v, err=%v", compensation, err)
	}
}
