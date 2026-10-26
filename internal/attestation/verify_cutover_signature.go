package attestation

import (
	"crypto/subtle"
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

const maximumCutoverReceiptClockSkew = 5 * time.Minute

func verifyCutoverDetachedSignature(
	manifestDocument []byte,
	signed cutoverv1.CredentialLeaseSignature,
	root TrustRoot,
	issuedAt time.Time,
) (string, error) {
	digest, err := DigestHex(manifestDocument)
	if err != nil {
		return "", err
	}
	if subtle.ConstantTimeCompare([]byte(digest), []byte(signed.ManifestSHA256)) != 1 {
		return "", fmt.Errorf("cutover credential lease digest does not match signature")
	}
	if err := root.Verify(manifestDocument, cutoverSignature(signed.Signature), issuedAt); err != nil {
		return "", fmt.Errorf("verify cutover credential lease signature: %w", err)
	}
	return digest, nil
}

func verifyCutoverEnvelope(
	envelope any,
	signature cutoverv1.Signature,
	issuedAt time.Time,
	root TrustRoot,
) error {
	document, err := MarshalCanonical(envelope)
	if err != nil {
		return err
	}
	if err := root.Verify(document, cutoverSignature(signature), issuedAt); err != nil {
		return fmt.Errorf("verify cutover receipt signature: %w", err)
	}
	return nil
}

func cutoverSignature(signature cutoverv1.Signature) Signature {
	return Signature{
		Algorithm: signature.Algorithm,
		KeyID:     signature.KeyID,
		Value:     signature.Value,
	}
}

func parseCutoverTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("cutover timestamp is invalid")
	}
	return parsed.UTC(), nil
}

func validateCutoverReceiptClock(issuedAt, now time.Time) error {
	if issuedAt.After(now.Add(maximumCutoverReceiptClockSkew)) {
		return fmt.Errorf("cutover receipt issue time is in the future")
	}
	return nil
}
