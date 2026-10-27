package environmentrecoveryv1

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func TestArtifactPurposeSeparationRotationAndTamper(t *testing.T) {
	now := time.Date(2026, 7, 23, 2, 0, 0, 0, time.UTC)
	bundle, keys := recoveryBundleFixture(t, now)
	input := Input{Purpose: PurposeRecoveryManifest,
		EnvironmentID: "018f0000-0000-7000-8000-000000000001",
		OperationID:   "018f0000-0000-7000-8000-000000000002",
		ArtifactID:    "018f0000-0000-7000-8000-000000000003",
		Sequence:      1, IssuedAt: now, ExpiresAt: now.Add(time.Hour),
		Payload: []byte(`{"portability":"same_environment"}`)}
	artifact, err := Build(input, bundle, keys[recoverysigningv1.PurposeRecoveryManifest])
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Parse(artifact.CanonicalBytes(), bundle, now); err != nil {
		t.Fatal(err)
	}
	if _, err = Build(input, bundle, keys[recoverysigningv1.PurposeRecoveryControl]); err == nil {
		t.Fatal("cross-purpose signer was accepted")
	}
	tampered := artifact.CanonicalBytes()
	tampered[len(tampered)-2] ^= 1
	if _, err = Parse(tampered, bundle, now); err == nil {
		t.Fatal("tampered artifact was accepted")
	}
}

func recoveryBundleFixture(
	t *testing.T,
	now time.Time,
) (recoverysigningv1.Bundle, map[recoverysigningv1.Purpose]ed25519.PrivateKey) {
	t.Helper()
	purposes := []recoverysigningv1.Purpose{recoverysigningv1.PurposeRecoveryManifest,
		recoverysigningv1.PurposeRecoveryControl, recoverysigningv1.PurposeRestoreReceipt}
	materials := make([]recoverysigningv1.PublicMaterial, 0, len(purposes))
	keys := make(map[recoverysigningv1.Purpose]ed25519.PrivateKey, len(purposes))
	for _, purpose := range purposes {
		public, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		var encoded [32]byte
		copy(encoded[:], public)
		materials = append(materials, recoverysigningv1.PublicMaterial{Purpose: purpose, PublicKey: encoded})
		keys[purpose] = private
	}
	bundle, err := recoverysigningv1.Build("018f0000-0000-7000-8000-000000000001", now, materials)
	if err != nil {
		t.Fatal(err)
	}
	return bundle, keys
}
