package licenseauthority

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type parsedTrustedIssuer struct {
	name string
	trustedIssuer
}

func verifyKeyset(
	envelope keysetEnvelopeWire,
	roots map[string]trustedLicenseKey,
	now time.Time,
) (parsedTrustedIssuer, error) {
	if envelope.Schema != KeysetEnvelopeSchema || envelope.Algorithm != SignatureAlgorithm ||
		envelope.Canonicalization != Canonicalization || !keyIDPattern.MatchString(envelope.KeyID) ||
		!digestPattern.MatchString(envelope.PayloadSHA256) {
		return parsedTrustedIssuer{}, ErrInvalidTrustBundle
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(envelope.Payload, MaximumDocumentBytes)
	if err != nil || !bytes.Equal(envelope.Payload, canonical) {
		return parsedTrustedIssuer{}, ErrInvalidTrustBundle
	}
	digest := sha256.Sum256(canonical)
	if hex.EncodeToString(digest[:]) != envelope.PayloadSHA256 {
		return parsedTrustedIssuer{}, ErrInvalidTrustBundle
	}
	parsed, issuedAt, err := parseKeysetPayload(canonical, now)
	if err != nil {
		return parsedTrustedIssuer{}, err
	}
	root, found := roots[envelope.KeyID]
	if !found || !keyCanVerify(root, issuedAt, now) {
		return parsedTrustedIssuer{}, ErrUntrustedTrustRoot
	}
	signature, err := base64.RawURLEncoding.Strict().DecodeString(envelope.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize ||
		base64.RawURLEncoding.EncodeToString(signature) != envelope.Signature ||
		!ed25519.Verify(root.publicKey, keysetSignatureMessage(canonical), signature) {
		return parsedTrustedIssuer{}, ErrInvalidTrustSignature
	}
	return parsed, nil
}

func keysetSignatureMessage(canonical []byte) []byte {
	message := make([]byte, 0, len(keysetSignatureDomain)+len(canonical))
	message = append(message, keysetSignatureDomain...)
	return append(message, canonical...)
}
