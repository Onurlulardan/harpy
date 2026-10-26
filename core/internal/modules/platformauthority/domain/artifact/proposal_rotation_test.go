package artifact

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func TestSignerRotationEvidenceBindsOldSignerNewKeyAndBundleBytes(t *testing.T) {
	fixture := newArtifactFixture(t)
	bundleBytes := []byte(`{"generation":2}`)
	bundleDigest := sha256.Sum256(bundleBytes)
	digest := hex.EncodeToString(bundleDigest[:])
	stateValue := fixture.state.Value
	stateValue.ProposalSignerGeneration = 2
	stateValue.ProposalSignerKID = "proposal.v2"
	stateValue.RecoveryTrustBundleSHA256 = &digest
	state, err := CanonicalizeState(stateValue)
	if err != nil {
		t.Fatal(err)
	}
	previous := strings.Repeat("b", 64)
	evidence := &SignerRotationEvidence{
		PreviousSignerGeneration: 1, NewSignerKID: "proposal.v2",
		NewSignerPublicKeySHA256: strings.Repeat("c", 64),
		RecoveryTrustBundleBytes: stringPointerForArtifact(
			base64.RawURLEncoding.EncodeToString(bundleBytes)),
	}
	proposal := Proposal{
		Schema: ProposalSchema, Purpose: "proposal_signer_rotation",
		OperationID:   "55555555-5555-4555-8555-555555555555",
		EnvironmentID: fixtureEnvironmentID, DeploymentMode: ModeOnPremise,
		ScopeKind: ScopeEnvironment, ScopeID: fixtureEnvironmentID,
		RegistryBindingID: fixtureBindingID, Generation: 2,
		PreviousHeadSHA256: &previous, StateSHA256: state.SHA256,
		RecoveryTrustBundleSHA256: &digest,
		IssuedAt:                  fixtureNow.Format(time.RFC3339),
		ExpiresAt:                 fixtureNow.Add(10 * time.Minute).Format(time.RFC3339),
		SignerKID:                 "proposal.v1", SignerRotation: evidence,
	}
	if _, err := SignProposal(proposal, state, ExpiryOnline,
		fixture.proposalPrivate); err != nil {
		t.Fatalf("old active signer rotation was rejected: %v", err)
	}
	selfEndorsed := proposal
	selfEndorsed.SignerKID = "proposal.v2"
	if _, err := SignProposal(selfEndorsed, state, ExpiryOnline,
		fixture.proposalPrivate); err == nil {
		t.Fatal("new signer self-endorsement was accepted")
	}
	wrongKID := proposal
	wrongEvidence := *evidence
	wrongEvidence.NewSignerKID = "proposal.v3"
	wrongKID.SignerRotation = &wrongEvidence
	if _, err := SignProposal(wrongKID, state, ExpiryOnline,
		fixture.proposalPrivate); err == nil {
		t.Fatal("rotation KID not pinned by state was accepted")
	}
	missingBundle := proposal
	missingEvidence := *evidence
	missingEvidence.RecoveryTrustBundleBytes = nil
	missingBundle.SignerRotation = &missingEvidence
	if _, err := SignProposal(missingBundle, state, ExpiryOnline,
		fixture.proposalPrivate); err == nil {
		t.Fatal("on-premise rotation without bundle bytes was accepted")
	}
}

func stringPointerForArtifact(value string) *string { return &value }
