package attestation

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"time"
)

type Signer struct {
	keyID      string
	privateKey ed25519.PrivateKey
	notBefore  time.Time
	notAfter   time.Time
}

func ParseSigner(document []byte, root TrustRoot) (Signer, error) {
	var decoded signingKeyDocument
	if err := decodeStrict(document, &decoded); err != nil {
		return Signer{}, err
	}
	if decoded.ContractVersion != "attestation-signing-key.v1" ||
		decoded.Algorithm != algorithmEd25519 || !keyIDPattern.MatchString(decoded.KeyID) {
		return Signer{}, fmt.Errorf("attestation signing key contract is invalid")
	}
	seed, err := decodeBase64URLExact(decoded.Seed, ed25519.SeedSize)
	if err != nil {
		return Signer{}, fmt.Errorf("attestation signing seed is invalid")
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	trusted, exists := root.keys[decoded.KeyID]
	if !exists || !trusted.publicKey.Equal(privateKey.Public()) {
		return Signer{}, fmt.Errorf("attestation signing key is not trusted")
	}
	return Signer{
		keyID: decoded.KeyID, privateKey: privateKey,
		notBefore: trusted.notBefore, notAfter: trusted.notAfter,
	}, nil
}

func (signer Signer) KeyID() string { return signer.keyID }

func (signer Signer) SignAt(document []byte, issuedAt time.Time) (Signature, error) {
	if signer.keyID == "" || len(signer.privateKey) != ed25519.PrivateKeySize {
		return Signature{}, fmt.Errorf("attestation signer is not initialized")
	}
	if issuedAt.Before(signer.notBefore) || !issuedAt.Before(signer.notAfter) {
		return Signature{}, fmt.Errorf("attestation signing key is outside its validity window")
	}
	canonical, err := Canonicalize(document)
	if err != nil {
		return Signature{}, err
	}
	value := base64.RawURLEncoding.EncodeToString(ed25519.Sign(signer.privateKey, canonical))
	return Signature{Algorithm: algorithmEd25519, KeyID: signer.keyID, Value: value}, nil
}
