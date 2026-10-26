package packagecontract

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
)

type Attestation struct {
	ManifestSHA256 string
	Signature      []byte
	KeyID          string
}

func Verify(
	document []byte,
	attestation Attestation,
	key ed25519.PublicKey,
) (Manifest, error) {
	if len(key) != ed25519.PublicKeySize || len(attestation.Signature) != ed25519.SignatureSize ||
		attestation.KeyID == "" || len(attestation.KeyID) > 128 ||
		!hashPattern.MatchString(attestation.ManifestSHA256) {
		return Manifest{}, ErrManifestTampered
	}
	digest := sha256.Sum256(document)
	if hex.EncodeToString(digest[:]) != attestation.ManifestSHA256 ||
		!ed25519.Verify(key, document, attestation.Signature) {
		return Manifest{}, ErrManifestTampered
	}
	manifest, err := Decode(document)
	if err != nil {
		return Manifest{}, ErrManifestTampered
	}
	return manifest, nil
}
