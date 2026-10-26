package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
)

func TestProtectedChangeAndSignerRotationAreTwoStepAndCrashSafe(t *testing.T) {
	fixture := newCoordinatorFixture(t)
	initializeFixtureAuthority(t, fixture)
	changed := fixture.state
	adminDigest := strings.Repeat("d", 64)
	changed.PrimaryAdminBindingSetSHA256 = &adminDigest
	prepared, err := fixture.service.PrepareChange(context.Background(), api.ProtectedChangeRequest{
		OperationID: "55555555-5555-4555-8555-555555555555",
		Purpose:     api.ChangeProtectedAdmin, State: changed,
		IssuedAt: testNow, Policy: api.ExpiryOnline,
	})
	if err != nil || fixture.registry.exchanges != 1 || prepared.Generation != 2 {
		t.Fatalf("prepare protected change = %+v, exchanges=%d, err=%v",
			prepared, fixture.registry.exchanges, err)
	}
	if _, err := fixture.service.AcknowledgeChange(
		context.Background(), prepared.OperationID,
	); err != nil {
		t.Fatal(err)
	}
	fixture.state = changed
	rotation := rotationRequest(t, fixture)
	prepared, err = fixture.service.PrepareSignerRotation(context.Background(), rotation)
	if err != nil || fixture.registry.exchanges != 2 || prepared.Generation != 3 {
		t.Fatalf("prepare signer rotation = %+v, exchanges=%d, err=%v",
			prepared, fixture.registry.exchanges, err)
	}
	fixture.store.failActivationOnce = true
	if _, err := fixture.service.AcknowledgeSignerRotation(
		context.Background(), prepared.OperationID,
	); err == nil {
		t.Fatal("post-ack signer activation crash was hidden")
	}
	receipt, err := fixture.service.AcknowledgeSignerRotation(
		context.Background(), prepared.OperationID,
	)
	if err != nil || receipt.Generation != 3 || fixture.registry.exchanges != 3 {
		t.Fatalf("signer activation replay = %+v, exchanges=%d, err=%v",
			receipt, fixture.registry.exchanges, err)
	}
	binding, _ := fixture.store.LoadBinding(context.Background(), testBindingID)
	if binding.SignerGeneration != 2 || binding.SignerKID != rotation.State.ProposalSignerKID ||
		binding.SignerPublicKeySHA256 != rotation.NewSignerPublicKeySHA256 ||
		binding.RecoveryTrustBundleSHA256 == nil ||
		*binding.RecoveryTrustBundleSHA256 != *rotation.State.RecoveryTrustBundleSHA256 ||
		binding.SignerStatus != "active" {
		t.Fatalf("rotated binding = %+v", binding)
	}
}

func TestSignerFenceBlocksPendingMutationAndOldSigner(t *testing.T) {
	fixture := newCoordinatorFixture(t)
	initializeFixtureAuthority(t, fixture)
	changed := fixture.state
	adminDigest := strings.Repeat("e", 64)
	changed.PrimaryAdminBindingSetSHA256 = &adminDigest
	prepared, err := fixture.service.PrepareChange(context.Background(), api.ProtectedChangeRequest{
		OperationID: "77777777-7777-4777-8777-777777777777",
		Purpose:     api.ChangeProtectedAdmin, State: changed,
		IssuedAt: testNow, Policy: api.ExpiryOnline,
	})
	if err != nil {
		t.Fatal(err)
	}
	binding, _ := fixture.store.LoadBinding(context.Background(), testBindingID)
	err = fixture.service.FenceSigner(context.Background(), api.SignerFenceRequest{
		RegistryBindingID:        testBindingID,
		ExpectedSignerGeneration: binding.SignerGeneration,
		ExpectedSignerKID:        binding.SignerKID, EvidenceSHA256: strings.Repeat("f", 64),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.AcknowledgeChange(
		context.Background(), prepared.OperationID,
	); !errors.Is(err, api.ErrConflict) {
		t.Fatalf("fenced pending mutation was accepted: %v", err)
	}
	if _, err := fixture.service.PrepareChange(context.Background(), api.ProtectedChangeRequest{
		OperationID: "88888888-8888-4888-8888-888888888888",
		Purpose:     api.ChangeProtectedAdmin, State: changed,
		IssuedAt: testNow, Policy: api.ExpiryOnline,
	}); !errors.Is(err, api.ErrConflict) {
		t.Fatalf("fenced signer prepared a new mutation: %v", err)
	}
}

func initializeFixtureAuthority(t *testing.T, fixture coordinatorFixture) {
	t.Helper()
	prepared, err := fixture.service.PrepareInitial(context.Background(), api.InitialAuthorityRequest{
		OperationID: testOperationID, State: fixture.state,
		IssuedAt: testNow, Policy: api.ExpiryOnline,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.service.AcknowledgePrepared(
		context.Background(), prepared.OperationID,
	); err != nil {
		t.Fatal(err)
	}
}

func rotationRequest(t *testing.T, fixture coordinatorFixture) api.SignerRotationRequest {
	t.Helper()
	newControl := testKeyPublic(10)
	bundle, err := recoveryv1.Rotate(fixture.recovery, testNow,
		[]recoveryv1.PublicMaterial{
			{Purpose: recoveryv1.PurposeRecoveryManifest, PublicKey: keyArray(testKeyPublic(9))},
			{Purpose: recoveryv1.PurposeRecoveryControl, PublicKey: keyArray(newControl)},
			{Purpose: recoveryv1.PurposeRestoreReceipt, PublicKey: keyArray(testKeyPublic(11))},
		})
	if err != nil {
		t.Fatal(err)
	}
	newKID, _ := bundle.KeyID(recoveryv1.PurposeRecoveryControl)
	digest := bundle.SHA256()
	publicDigest := sha256.Sum256(newControl)
	state := fixture.state
	state.ProposalSignerGeneration = 2
	state.ProposalSignerKID = newKID
	state.RecoveryTrustBundleSHA256 = &digest
	return api.SignerRotationRequest{
		OperationID: "66666666-6666-4666-8666-666666666666", State: state,
		NewSignerPublicKeySHA256: hex.EncodeToString(publicDigest[:]),
		RecoveryTrustBundle:      bundle.CanonicalBytes(), IssuedAt: testNow,
		Policy: api.ExpiryOnline,
	}
}
