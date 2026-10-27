package environmentrecoveryv1

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func Build(
	input Input,
	bundle recoverysigningv1.Bundle,
	privateKey ed25519.PrivateKey,
) (Artifact, error) {
	purpose, valid := signingPurpose(input.Purpose)
	payload, err := canonicaljsonv1.CanonicalizeObject(input.Payload, MaximumPayloadBytes)
	if err != nil || !valid || !validInput(input) || bundle.EnvironmentID() != input.EnvironmentID ||
		len(privateKey) != ed25519.PrivateKeySize {
		return Artifact{}, fmt.Errorf("recovery artifact input is invalid")
	}
	public, found := bundle.PublicKey(purpose)
	if !found || !privateKey.Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(public[:])) {
		return Artifact{}, fmt.Errorf("recovery artifact signer is invalid")
	}
	keyID, found := bundle.KeyID(purpose)
	if !found {
		return Artifact{}, fmt.Errorf("recovery artifact signer is invalid")
	}
	payloadDigest := sha256.Sum256(payload)
	document := wire{Schema: Schema, Purpose: input.Purpose,
		EnvironmentID: input.EnvironmentID, OperationID: input.OperationID,
		ArtifactID: input.ArtifactID, Sequence: input.Sequence,
		IssuedAt: input.IssuedAt.Format(timeFormat), ExpiresAt: input.ExpiresAt.Format(timeFormat),
		PayloadSHA256: hex.EncodeToString(payloadDigest[:]), Payload: payload,
		KeyGeneration: bundle.Generation(), SignerKID: keyID, Algorithm: Algorithm}
	unsigned, err := canonicalWire(document)
	if err != nil {
		return Artifact{}, err
	}
	domain := signatureDomain(input.Purpose, unsigned)
	document.Signature = base64.RawURLEncoding.EncodeToString(ed25519.Sign(privateKey, domain))
	signed, err := canonicalWire(document)
	if err != nil {
		return Artifact{}, err
	}
	return parseVerified(signed, bundle, input.IssuedAt)
}

func canonicalWire(value wire) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("recovery artifact encoding failed")
	}
	return canonicaljsonv1.CanonicalizeObject(raw, MaximumArtifactBytes)
}
