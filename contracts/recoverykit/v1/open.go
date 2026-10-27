package recoverykitv1

import (
	"crypto/ecdh"
	"crypto/hpke"
	"fmt"
	"time"
)

func Open(
	document []byte,
	claimPrivateKey [32]byte,
	expected OpenExpected,
) (Kit, error) {
	var wire sealedWire
	if err := parseCanonical(document, &wire); err != nil ||
		!validSealedWire(wire, claimPrivateKey, expected) {
		return Kit{}, fmt.Errorf("sealed recovery kit is invalid")
	}
	encapsulated, encValid := decode(wire.EncapsulatedKey, 32)
	ciphertext, ciphertextValid := decode(wire.Ciphertext, -1)
	if !encValid || !ciphertextValid || len(ciphertext) < 17 {
		clear(encapsulated)
		clear(ciphertext)
		return Kit{}, fmt.Errorf("sealed recovery kit ciphertext is invalid")
	}
	private, err := ecdh.X25519().NewPrivateKey(claimPrivateKey[:])
	if err != nil {
		return Kit{}, fmt.Errorf("recovery kit claim key is invalid")
	}
	hpkePrivate, err := hpke.NewDHKEMPrivateKey(private)
	if err != nil {
		return Kit{}, fmt.Errorf("recovery kit claim key is invalid")
	}
	metadata := metadataFromWire(wire)
	aad, err := canonicalObject(metadata)
	if err != nil {
		return Kit{}, err
	}
	recipient, err := hpke.NewRecipient(
		encapsulated, hpkePrivate, hpke.HKDFSHA256(), hpke.AES256GCM(), hpkeInfo(aad))
	clear(encapsulated)
	if err != nil {
		clear(ciphertext)
		return Kit{}, fmt.Errorf("recovery kit HPKE recipient failed")
	}
	plaintext, err := recipient.Open(aad, ciphertext)
	clear(ciphertext)
	if err != nil {
		return Kit{}, fmt.Errorf("recovery kit HPKE open failed")
	}
	defer clear(plaintext)
	kit, err := Parse(plaintext)
	if err != nil || kit.EnvironmentID() != expected.EnvironmentID ||
		kit.OperationID() != expected.OperationID || kit.Generation() != expected.Generation {
		return Kit{}, fmt.Errorf("recovery kit plaintext binding is invalid")
	}
	return kit, nil
}

func validSealedWire(wire sealedWire, privateKey [32]byte, expected OpenExpected) bool {
	issuedAt, issueErr := time.Parse(canonicalTimeFormat, wire.IssuedAt)
	expiresAt, expiryErr := time.Parse(canonicalTimeFormat, wire.ExpiresAt)
	private, keyErr := ecdh.X25519().NewPrivateKey(privateKey[:])
	return issueErr == nil && expiryErr == nil && keyErr == nil &&
		!expected.Now.IsZero() && expected.Now.Location() == time.UTC &&
		canonicalTime(issuedAt) && canonicalTime(expiresAt) && !expected.Now.Before(issuedAt) &&
		expected.Now.Before(expiresAt) && wire.Schema == SealedSchema && wire.Suite == Suite &&
		wire.EnvironmentID == expected.EnvironmentID && wire.OperationID == expected.OperationID &&
		wire.Generation == expected.Generation && validUUID(wire.EnvironmentID) &&
		validUUID(wire.OperationID) && wire.Generation > 0 &&
		wire.ClaimPublicKeySHA256 == digestHex(private.PublicKey().Bytes())
}

func metadataFromWire(wire sealedWire) sealedMetadata {
	return sealedMetadata{Schema: wire.Schema, Suite: wire.Suite,
		EnvironmentID: wire.EnvironmentID, OperationID: wire.OperationID,
		Generation: wire.Generation, ClaimPublicKeySHA256: wire.ClaimPublicKeySHA256,
		IssuedAt: wire.IssuedAt, ExpiresAt: wire.ExpiresAt}
}
