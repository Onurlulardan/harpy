package attestation

import (
	"crypto/subtle"
	"fmt"
	"time"
)

type manifestSignatureDocument struct {
	ContractVersion string    `json:"contract_version"`
	ManifestSHA256  string    `json:"manifest_sha256"`
	Signature       Signature `json:"signature"`
}

func verifyManifestSignature(
	manifestDocument, signatureDocument []byte,
	expectedContract string,
	root TrustRoot,
	issuedAt time.Time,
) (string, error) {
	var signed manifestSignatureDocument
	if err := decodeStrict(signatureDocument, &signed); err != nil {
		return "", fmt.Errorf("verify manifest signature contract: %w", err)
	}
	if signed.ContractVersion != expectedContract ||
		!digestPattern.MatchString(signed.ManifestSHA256) {
		return "", fmt.Errorf("manifest signature contract is invalid")
	}
	digest, err := DigestHex(manifestDocument)
	if err != nil {
		return "", err
	}
	if subtle.ConstantTimeCompare([]byte(digest), []byte(signed.ManifestSHA256)) != 1 {
		return "", fmt.Errorf("manifest digest does not match signature")
	}
	if err := root.Verify(manifestDocument, signed.Signature, issuedAt); err != nil {
		return "", fmt.Errorf("verify manifest signature: %w", err)
	}
	return digest, nil
}
