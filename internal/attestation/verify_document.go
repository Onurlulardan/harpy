package attestation

import (
	"crypto/ed25519"
	"fmt"
	"time"
)

func (root TrustRoot) Verify(document []byte, signature Signature, issuedAt time.Time) error {
	if signature.Algorithm != algorithmEd25519 || !keyIDPattern.MatchString(signature.KeyID) {
		return fmt.Errorf("attestation signature identity is invalid")
	}
	trusted, exists := root.keys[signature.KeyID]
	if !exists {
		return fmt.Errorf("attestation signature key is not trusted")
	}
	if issuedAt.Before(trusted.notBefore) || !issuedAt.Before(trusted.notAfter) {
		return fmt.Errorf("attestation signature key is outside its validity window")
	}
	value, err := decodeBase64URLExact(signature.Value, ed25519.SignatureSize)
	if err != nil {
		return fmt.Errorf("attestation signature value is invalid")
	}
	canonical, err := Canonicalize(document)
	if err != nil {
		return err
	}
	if !ed25519.Verify(trusted.publicKey, canonical, value) {
		return fmt.Errorf("attestation signature verification failed")
	}
	return nil
}
