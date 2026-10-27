package artifact

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

var fixtureNow = time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)

const (
	fixtureEnvironmentID = "11111111-1111-4111-8111-111111111111"
	fixtureBindingID     = "22222222-2222-4222-8222-222222222222"
	fixtureOperationID   = "33333333-3333-4333-8333-333333333333"
	fixtureOwnerID       = "44444444-4444-4444-8444-444444444444"
)

type artifactFixture struct {
	releasePublic, proposalPublic, registryPublic    ed25519.PublicKey
	releasePrivate, proposalPrivate, registryPrivate ed25519.PrivateKey
	state                                            CanonicalState
	proposal                                         SignedProposal
	trust                                            VerifiedTrustBundle
}

func newArtifactFixture(t *testing.T) artifactFixture {
	t.Helper()
	releasePublic, releasePrivate := deterministicKey(1)
	proposalPublic, proposalPrivate := deterministicKey(2)
	registryPublic, registryPrivate := deterministicKey(3)
	adminDigest := strings.Repeat("a", 64)
	ownerGeneration := int64(1)
	owner := fixtureOwnerID
	state, err := CanonicalizeState(State{
		Schema: StateSchema, EnvironmentID: fixtureEnvironmentID,
		DeploymentMode: ModeOnPremise, RegistryBindingID: fixtureBindingID,
		ScopeKind: ScopeEnvironment, ScopeID: fixtureEnvironmentID,
		ProposalSignerGeneration: 1, ProposalSignerKID: "proposal.v1",
		OwnerBindingID: &owner, OwnerBindingGeneration: &ownerGeneration,
		PrimaryAdminBindingSetSHA256: &adminDigest,
	})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := SignProposal(Proposal{
		Schema: ProposalSchema, Purpose: "initial_environment_authority",
		OperationID: fixtureOperationID, EnvironmentID: fixtureEnvironmentID,
		DeploymentMode: ModeOnPremise, ScopeKind: ScopeEnvironment,
		ScopeID: fixtureEnvironmentID, RegistryBindingID: fixtureBindingID,
		Generation: 1, StateSHA256: state.SHA256,
		IssuedAt:  fixtureNow.Format(time.RFC3339),
		ExpiresAt: fixtureNow.Add(10 * time.Minute).Format(time.RFC3339),
		SignerKID: "proposal.v1",
	}, state, ExpiryOnline, proposalPrivate)
	if err != nil {
		t.Fatal(err)
	}
	bundleBytes, err := SignTrustBundle(TrustBundle{
		Schema: TrustBundleSchema, RegistryKind: RegistryDevelopment, TrustGeneration: 1,
		Active: TrustKey{
			KID: "registry.v1", Algorithm: AlgorithmEd25519,
			PublicKey: base64.RawURLEncoding.EncodeToString(registryPublic),
			NotBefore: fixtureNow.Add(-time.Hour).Format(time.RFC3339),
			SignUntil: fixtureNow.Add(time.Hour).Format(time.RFC3339),
		}, Retired: []RetiredTrustKey{},
	}, releasePrivate)
	if err != nil {
		t.Fatal(err)
	}
	trust, err := VerifyTrustBundle(
		bundleBytes, RegistryDevelopment, fixtureNow, releasePublic,
	)
	if err != nil {
		t.Fatal(err)
	}
	return artifactFixture{
		releasePublic: releasePublic, releasePrivate: releasePrivate,
		proposalPublic: proposalPublic, proposalPrivate: proposalPrivate,
		registryPublic: registryPublic, registryPrivate: registryPrivate,
		state: state, proposal: proposal, trust: trust,
	}
}

func deterministicKey(marker byte) (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := bytesOf(marker, ed25519.SeedSize)
	private := ed25519.NewKeyFromSeed(seed)
	return private.Public().(ed25519.PublicKey), private
}

func bytesOf(value byte, count int) []byte {
	result := make([]byte, count)
	for index := range result {
		result[index] = value
	}
	return result
}

func (fixture artifactFixture) ack(t *testing.T, private ed25519.PrivateKey) VerifiedAck {
	t.Helper()
	ack, err := SignAck(Ack{
		Schema: AckSchema, RegistryKind: RegistryDevelopment,
		RegistryBindingID: fixtureBindingID, ScopeKind: ScopeEnvironment,
		ScopeID: fixtureEnvironmentID, RegistrySequence: 1,
		OperationID: fixtureOperationID, ProposalSHA256: fixture.proposal.SHA256,
		Generation: 1, StateSHA256: fixture.state.SHA256,
		IssuedAt: fixtureNow.Format(time.RFC3339), RegistryKID: "registry.v1",
	}, fixture.proposal, private)
	if err != nil {
		t.Fatal(err)
	}
	return ack
}
