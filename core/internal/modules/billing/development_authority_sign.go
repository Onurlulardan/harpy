package billing

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
)

func canonicalDevelopmentLicense(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(raw, 64*1024)
}

func signDevelopmentKeyset(
	payload licensev2.KeysetPayload,
	privateKey ed25519.PrivateKey,
) (licensev2.KeysetEnvelope, error) {
	canonical, err := canonicalDevelopmentLicense(payload)
	if err != nil {
		return licensev2.KeysetEnvelope{}, err
	}
	digest := sha256.Sum256(canonical)
	message := append([]byte(licensev2.KeysetEnvelopeSchema+"\x00"), canonical...)
	signature := ed25519.Sign(privateKey, message)
	clear(message)
	return licensev2.KeysetEnvelope{
		Schema: licensev2.KeysetEnvelopeSchema, KeyID: developmentRootKeyID,
		Algorithm: licensev2.SignatureAlgorithm, Canonicalization: licensev2.Canonicalization,
		PayloadSHA256: hex.EncodeToString(digest[:]), Payload: canonical,
		Signature: base64.RawURLEncoding.EncodeToString(signature),
	}, nil
}

func signDevelopmentSnapshot(
	payload licensev2.SnapshotPayload,
	privateKey ed25519.PrivateKey,
) ([]byte, error) {
	canonical, err := canonicalDevelopmentLicense(payload)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(canonical)
	message := append([]byte(licensev2.SnapshotEnvelopeSchema+"\x00"), canonical...)
	signature := ed25519.Sign(privateKey, message)
	clear(message)
	envelope := licensev2.SnapshotEnvelope{
		Schema: licensev2.SnapshotEnvelopeSchema, KeyID: developmentSignerKeyID,
		Algorithm: licensev2.SignatureAlgorithm, Canonicalization: licensev2.Canonicalization,
		PayloadSHA256: hex.EncodeToString(digest[:]), Payload: canonical,
		Signature: base64.RawURLEncoding.EncodeToString(signature),
	}
	return canonicalDevelopmentLicense(envelope)
}
