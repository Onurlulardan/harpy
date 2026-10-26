package environmentrecoveryv1

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

func Parse(
	document []byte,
	bundle recoverysigningv1.Bundle,
	now time.Time,
) (Artifact, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumArtifactBytes)
	if err != nil || !bytes.Equal(canonical, document) {
		return Artifact{}, fmt.Errorf("recovery artifact is invalid")
	}
	return parseVerified(canonical, bundle, now)
}

func parseVerified(
	canonical []byte,
	bundle recoverysigningv1.Bundle,
	now time.Time,
) (Artifact, error) {
	var document wire
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&document) != nil || requireEOF(decoder) != nil ||
		!validWire(document, bundle, now) {
		return Artifact{}, fmt.Errorf("recovery artifact is invalid")
	}
	purpose, _ := signingPurpose(document.Purpose)
	public, generation, found := bundle.VerificationKey(purpose, document.SignerKID)
	if !found || generation != document.KeyGeneration {
		return Artifact{}, fmt.Errorf("recovery artifact signer is invalid")
	}
	signature, err := base64.RawURLEncoding.DecodeString(document.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return Artifact{}, fmt.Errorf("recovery artifact signature is invalid")
	}
	signedWire := document
	signedWire.Signature = ""
	unsigned, err := canonicalWire(signedWire)
	if err != nil || !ed25519.Verify(public[:], signatureDomain(document.Purpose, unsigned), signature) {
		return Artifact{}, fmt.Errorf("recovery artifact signature is invalid")
	}
	payloadDigest := sha256.Sum256(document.Payload)
	if hex.EncodeToString(payloadDigest[:]) != document.PayloadSHA256 {
		return Artifact{}, fmt.Errorf("recovery artifact payload is invalid")
	}
	issuedAt, _ := time.Parse(timeFormat, document.IssuedAt)
	expiresAt, _ := time.Parse(timeFormat, document.ExpiresAt)
	digest := sha256.Sum256(canonical)
	return Artifact{wire: document, canonical: append([]byte(nil), canonical...),
		digest: hex.EncodeToString(digest[:]), issuedAt: issuedAt, expiresAt: expiresAt}, nil
}
