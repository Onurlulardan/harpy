package recoverykitv1

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func Build(input BuildInput) (Kit, error) {
	if !validBuildInput(input) {
		return Kit{}, fmt.Errorf("recovery kit input is invalid")
	}
	recoveryKeys := make([]recoveryKeyWire, len(input.RecoveryKeys))
	for index, key := range input.RecoveryKeys {
		recoveryKeys[index] = recoveryKeyWire{KeyID: key.KeyID, Algorithm: "X25519",
			PrivateKey: encode(key.PrivateKey[:]), State: key.State}
	}
	signingKeys := make([]signingKeyWire, len(input.SigningKeys))
	for index, key := range input.SigningKeys {
		signingKeys[index] = signingKeyWire{Purpose: key.Purpose, KeyID: key.KeyID,
			Algorithm: "Ed25519", PrivateKey: encode(key.PrivateKey)}
	}
	wire := plaintextWire{Schema: PlaintextSchema,
		EnvironmentID: input.EnvironmentID, OperationID: input.OperationID,
		Generation:              input.Generation,
		DeploymentBindingSHA256: hex.EncodeToString(input.DeploymentBindingSHA256[:]),
		RecoveryKeys:            recoveryKeys, SigningKeys: signingKeys,
		RecoveryTrustBundleBase64URL: base64.RawURLEncoding.EncodeToString(
			input.TrustBundle.CanonicalBytes()),
		AckProofBase64URL: encode(input.AckProof[:]),
		PreviousKitSHA256: optionalDigest(input.PreviousKitSHA256),
		GeneratedAt:       input.GeneratedAt.Format(canonicalTimeFormat)}
	canonical, err := canonicalObject(wire)
	if err != nil {
		return Kit{}, err
	}
	return Kit{wire: wire, canonical: canonical}, nil
}

func validBuildInput(input BuildInput) bool {
	if !validUUID(input.EnvironmentID) || !validUUID(input.OperationID) ||
		input.Generation < 1 || input.DeploymentBindingSHA256 == [32]byte{} ||
		input.AckProof == [32]byte{} || !canonicalTime(input.GeneratedAt) ||
		input.TrustBundle.EnvironmentID() != input.EnvironmentID ||
		input.TrustBundle.Generation() != input.Generation ||
		!validRecoveryKeys(input.RecoveryKeys) ||
		!validSigningKeys(input.SigningKeys, input.TrustBundle) {
		return false
	}
	return input.PreviousKitSHA256 == nil || *input.PreviousKitSHA256 != [32]byte{}
}

func validSigningKeys(keys []SigningKeyMaterial, bundle recoverysigningv1.Bundle) bool {
	want := []recoverysigningv1.Purpose{recoverysigningv1.PurposeRecoveryManifest,
		recoverysigningv1.PurposeRecoveryControl, recoverysigningv1.PurposeRestoreReceipt}
	if len(keys) != len(want) {
		return false
	}
	for index, key := range keys {
		public, found := bundle.PublicKey(want[index])
		keyID, idFound := bundle.KeyID(want[index])
		if !found || !idFound || key.Purpose != want[index] || key.KeyID != keyID ||
			len(key.PrivateKey) != ed25519.PrivateKeySize ||
			!key.PrivateKey.Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(public[:])) {
			return false
		}
	}
	return true
}

func optionalDigest(value *[32]byte) *string {
	if value == nil {
		return nil
	}
	encoded := hex.EncodeToString(value[:])
	return &encoded
}
