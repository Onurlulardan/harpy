package recoverykitv1

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ed25519"
	"testing"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func TestSealOpensFixedPlaintextUnderHarpyDomain(t *testing.T) {
	const harpyHPKEInfoDomain = "harpy.environment-recovery-kit.hpke.v1\x00"
	if hpkeInfoDomain != harpyHPKEInfoDomain ||
		!bytes.HasPrefix(hpkeInfo(nil), []byte(harpyHPKEInfoDomain)) {
		t.Fatalf("HPKE info domain must be the Harpy domain, got %q", hpkeInfoDomain)
	}
	now := time.Date(2026, 7, 23, 6, 0, 0, 0, time.UTC)
	bundle, signing := fixedSigningMaterial(t, now)
	claimSeed := fixedBytes(0x21)
	claimPrivate, err := ecdh.X25519().NewPrivateKey(claimSeed[:])
	if err != nil {
		t.Fatal(err)
	}
	var claimPublic, claimSecret [32]byte
	copy(claimPublic[:], claimPrivate.PublicKey().Bytes())
	copy(claimSecret[:], claimPrivate.Bytes())
	kit, err := Build(BuildInput{EnvironmentID: kitEnvironmentID, OperationID: kitOperationID,
		Generation: 1, DeploymentBindingSHA256: fixedBytes(0x01), AckProof: fixedBytes(0x02),
		RecoveryKeys: []RecoveryKeyMaterial{{KeyID: "recovery-key-1",
			PrivateKey: fixedBytes(0x11), State: "active"}},
		SigningKeys: signing, TrustBundle: bundle, GeneratedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	plaintext := kit.CanonicalBytes()
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
	if !bytes.Equal(opened.CanonicalBytes(), plaintext) {
		t.Fatal("opened recovery kit plaintext does not match the fixed plaintext")
	}
}

func fixedBytes(fill byte) [32]byte {
	var value [32]byte
	for index := range value {
		value[index] = fill
	}
	return value
}

func fixedSigningMaterial(
	t *testing.T, now time.Time,
) (recoverysigningv1.Bundle, []SigningKeyMaterial) {
	t.Helper()
	purposes := []recoverysigningv1.Purpose{recoverysigningv1.PurposeRecoveryManifest,
		recoverysigningv1.PurposeRecoveryControl, recoverysigningv1.PurposeRestoreReceipt}
	public := make([]recoverysigningv1.PublicMaterial, len(purposes))
	private := make([]ed25519.PrivateKey, len(purposes))
	for index, purpose := range purposes {
		seed := fixedBytes(byte(0x31 + index))
		key := ed25519.NewKeyFromSeed(seed[:])
		copy(public[index].PublicKey[:], key.Public().(ed25519.PublicKey))
		public[index].Purpose, private[index] = purpose, key
	}
	bundle, err := recoverysigningv1.Build(kitEnvironmentID, now, public)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]SigningKeyMaterial, len(purposes))
	for index, purpose := range purposes {
		keyID, found := bundle.KeyID(purpose)
		if !found {
			t.Fatalf("signing key id missing for %s", purpose)
		}
		result[index] = SigningKeyMaterial{Purpose: purpose, KeyID: keyID, PrivateKey: private[index]}
	}
	return bundle, result
}
