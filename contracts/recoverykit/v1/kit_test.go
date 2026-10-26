package recoverykitv1

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func TestRecoveryKitHPKERoundTripAndBinding(t *testing.T) {
	now := time.Date(2026, 7, 23, 4, 0, 0, 0, time.UTC)
	bundle, signing := kitSigningFixture(t, now)
	recoveryPrivate, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	claimPrivate, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var recoveryKey, claimPublic, claimSecret [32]byte
	copy(recoveryKey[:], recoveryPrivate.Bytes())
	copy(claimPublic[:], claimPrivate.PublicKey().Bytes())
	copy(claimSecret[:], claimPrivate.Bytes())
	input := BuildInput{EnvironmentID: kitEnvironmentID, OperationID: kitOperationID,
		Generation: 1, DeploymentBindingSHA256: [32]byte{1}, AckProof: [32]byte{2},
		RecoveryKeys: []RecoveryKeyMaterial{{KeyID: "recovery-key-1",
			PrivateKey: recoveryKey, State: "active"}},
		SigningKeys: signing, TrustBundle: bundle, GeneratedAt: now}
	kit, err := Build(input)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := Seal(kit, SealInput{EnvironmentID: kitEnvironmentID,
		OperationID: kitOperationID, Generation: 1, ClaimPublicKey: claimPublic,
		IssuedAt: now, ExpiresAt: now.Add(15 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	opened, err := Open(sealed, claimSecret, OpenExpected{EnvironmentID: kitEnvironmentID,
		OperationID: kitOperationID, Generation: 1, Now: now.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if opened.AckProofBase64URL() != "AgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" {
		t.Fatalf("ack proof did not survive the sealed kit: %s", opened.AckProofBase64URL())
	}
	wrong := claimSecret
	wrong[10] ^= 1
	if _, err = Open(sealed, wrong, OpenExpected{EnvironmentID: kitEnvironmentID,
		OperationID: kitOperationID, Generation: 1, Now: now.Add(time.Minute)}); err == nil {
		t.Fatal("wrong claim private key opened the recovery kit")
	}
}

const (
	kitEnvironmentID = "018f0000-0000-7000-8000-000000000001"
	kitOperationID   = "018f0000-0000-7000-8000-000000000002"
)

func kitSigningFixture(
	t *testing.T, now time.Time,
) (recoverysigningv1.Bundle, []SigningKeyMaterial) {
	t.Helper()
	purposes := []recoverysigningv1.Purpose{recoverysigningv1.PurposeRecoveryManifest,
		recoverysigningv1.PurposeRecoveryControl, recoverysigningv1.PurposeRestoreReceipt}
	public := make([]recoverysigningv1.PublicMaterial, len(purposes))
	private := make([]ed25519.PrivateKey, len(purposes))
	for index, purpose := range purposes {
		key, secret, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		copy(public[index].PublicKey[:], key)
		public[index].Purpose, private[index] = purpose, secret
	}
	bundle, err := recoverysigningv1.Build(kitEnvironmentID, now, public)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]SigningKeyMaterial, len(purposes))
	for index, purpose := range purposes {
		keyID, _ := bundle.KeyID(purpose)
		result[index] = SigningKeyMaterial{Purpose: purpose, KeyID: keyID,
			PrivateKey: private[index]}
	}
	return bundle, result
}
