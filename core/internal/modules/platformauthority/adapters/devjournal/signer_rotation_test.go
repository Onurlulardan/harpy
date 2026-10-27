package devjournal

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func TestJournalSignerRotationPinsNewKeyAndRetiresOldSigner(t *testing.T) {
	fixture := newJournalFixture(t)
	ctx := context.Background()
	if err := fixture.registry.EnsureBinding(ctx, fixture.proof); err != nil {
		t.Fatal(err)
	}
	initial := fixture.signProposal(t, journalOperation, 1, nil)
	initialAck, err := fixture.registry.ExchangeProposal(ctx, initial)
	if err != nil {
		t.Fatal(err)
	}
	verifiedInitial, err := artifact.VerifyAckSignature(initialAck,
		artifact.RegistryDevelopment,
		fixture.registry.privateKey.Public().(ed25519.PublicKey))
	if err != nil {
		t.Fatal(err)
	}
	newPublic, newPrivate := journalKey(9)
	rotated, err := recoveryv1.Rotate(fixture.recovery, fixture.now,
		journalRecoveryMaterial(newPublic, 10))
	if err != nil {
		t.Fatal(err)
	}
	newKID, _ := rotated.KeyID(recoveryv1.PurposeRecoveryControl)
	fixture.resolver[newKID] = newPublic
	newDigest := sha256.Sum256(newPublic)
	encodedBundle := base64.RawURLEncoding.EncodeToString(rotated.CanonicalBytes())
	rotatedState := fixture.state.Value
	rotatedState.ProposalSignerGeneration = 2
	rotatedState.ProposalSignerKID = newKID
	rotatedState.RecoveryTrustBundleSHA256 = stringPointer(rotated.SHA256())
	canonicalRotatedState, err := artifact.CanonicalizeState(rotatedState)
	if err != nil {
		t.Fatal(err)
	}
	rotation := signJournalRotation(t, fixture, canonicalRotatedState,
		verifiedInitial.SHA256, newKID, hex.EncodeToString(newDigest[:]), encodedBundle)
	rotationAckBytes, err := fixture.registry.ExchangeProposal(ctx, rotation)
	if err != nil {
		t.Fatal(err)
	}
	if replay, replayErr := fixture.registry.ExchangeProposal(ctx, rotation); replayErr != nil || string(replay) != string(rotationAckBytes) {
		t.Fatalf("rotation replay failed: %v", replayErr)
	}
	rotationAck, err := artifact.VerifyAckSignature(rotationAckBytes,
		artifact.RegistryDevelopment,
		fixture.registry.privateKey.Public().(ed25519.PublicKey))
	if err != nil {
		t.Fatal(err)
	}
	assertRetiredSignerRejected(t, fixture, rotationAck.SHA256)
	adminDigest := strings.Repeat("e", 64)
	rotatedState.PrimaryAdminBindingSetSHA256 = &adminDigest
	changedState, _ := artifact.CanonicalizeState(rotatedState)
	change := signJournalChange(t, fixture, changedState, rotationAck.SHA256,
		newKID, newPrivate)
	if _, err := fixture.registry.ExchangeProposal(ctx, change); err != nil {
		t.Fatalf("new active signer was rejected: %v", err)
	}
}

func signJournalRotation(
	t *testing.T,
	fixture journalFixture,
	state artifact.CanonicalState,
	previous, newKID, newDigest, encodedBundle string,
) artifact.SignedProposal {
	t.Helper()
	proposal, err := artifact.SignProposal(artifact.Proposal{
		Schema: artifact.ProposalSchema, Purpose: "proposal_signer_rotation",
		OperationID:   "88888888-8888-4888-8888-888888888888",
		EnvironmentID: journalEnvironment, DeploymentMode: artifact.ModeOnPremise,
		ScopeKind: artifact.ScopeEnvironment, ScopeID: journalEnvironment,
		RegistryBindingID: journalBindingID, Generation: 2,
		PreviousHeadSHA256: &previous, StateSHA256: state.SHA256,
		RecoveryTrustBundleSHA256: state.Value.RecoveryTrustBundleSHA256,
		IssuedAt:                  fixture.now.Format(time.RFC3339),
		ExpiresAt:                 fixture.now.Add(10 * time.Minute).Format(time.RFC3339),
		SignerKID:                 fixture.proof.Value.ControlKID,
		SignerRotation: &artifact.SignerRotationEvidence{
			PreviousSignerGeneration: 1, NewSignerKID: newKID,
			NewSignerPublicKeySHA256: newDigest,
			RecoveryTrustBundleBytes: &encodedBundle,
		},
	}, state, artifact.ExpiryOnline, fixture.proposalPrivate)
	if err != nil {
		t.Fatal(err)
	}
	return proposal
}

func assertRetiredSignerRejected(t *testing.T, fixture journalFixture, previous string) {
	t.Helper()
	adminDigest := strings.Repeat("e", 64)
	oldState := fixture.state.Value
	oldState.PrimaryAdminBindingSetSHA256 = &adminDigest
	canonical, _ := artifact.CanonicalizeState(oldState)
	proposal, err := artifact.SignProposal(artifact.Proposal{
		Schema: artifact.ProposalSchema, Purpose: "protected_admin_change",
		OperationID:   "77777777-7777-4777-8777-777777777777",
		EnvironmentID: journalEnvironment, DeploymentMode: artifact.ModeOnPremise,
		ScopeKind: artifact.ScopeEnvironment, ScopeID: journalEnvironment,
		RegistryBindingID: journalBindingID, Generation: 3,
		PreviousHeadSHA256: &previous, StateSHA256: canonical.SHA256,
		RecoveryTrustBundleSHA256: fixture.state.Value.RecoveryTrustBundleSHA256,
		IssuedAt:                  fixture.now.Format(time.RFC3339),
		ExpiresAt:                 fixture.now.Add(10 * time.Minute).Format(time.RFC3339),
		SignerKID:                 fixture.proof.Value.ControlKID,
	}, canonical, artifact.ExpiryOnline, fixture.proposalPrivate)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.registry.ExchangeProposal(context.Background(), proposal); err == nil {
		t.Fatal("retired signer was accepted")
	}
}

func signJournalChange(
	t *testing.T,
	fixture journalFixture,
	state artifact.CanonicalState,
	previous, kid string,
	privateKey ed25519.PrivateKey,
) artifact.SignedProposal {
	t.Helper()
	proposal, err := artifact.SignProposal(artifact.Proposal{
		Schema: artifact.ProposalSchema, Purpose: "protected_admin_change",
		OperationID:   "99999999-9999-4999-8999-999999999999",
		EnvironmentID: journalEnvironment, DeploymentMode: artifact.ModeOnPremise,
		ScopeKind: artifact.ScopeEnvironment, ScopeID: journalEnvironment,
		RegistryBindingID: journalBindingID, Generation: 3,
		PreviousHeadSHA256: &previous, StateSHA256: state.SHA256,
		RecoveryTrustBundleSHA256: state.Value.RecoveryTrustBundleSHA256,
		IssuedAt:                  fixture.now.Format(time.RFC3339),
		ExpiresAt:                 fixture.now.Add(10 * time.Minute).Format(time.RFC3339),
		SignerKID:                 kid,
	}, state, artifact.ExpiryOnline, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	return proposal
}
