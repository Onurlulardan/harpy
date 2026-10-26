package aicredentialv1

import (
	"crypto/ecdh"
	"encoding/base64"
)

type Opener struct{ privateKey *ecdh.PrivateKey }

func NewOpener(privateKey []byte) (*Opener, error) {
	parsed, err := ecdh.X25519().NewPrivateKey(privateKey)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	return &Opener{privateKey: parsed}, nil
}

func (opener *Opener) Open(document []byte, expected AAD) ([]byte, error) {
	if opener == nil || opener.privateKey == nil {
		return nil, ErrInvalidEnvelope
	}
	envelope, err := Parse(document)
	if err != nil || envelope.AAD != expected || expected.validate() != nil {
		return nil, ErrInvalidEnvelope
	}
	authenticated, err := aadBytes(envelope.AAD)
	if err != nil {
		return nil, err
	}
	ephemeralBytes, err := decodeExact(envelope.EphemeralPublicKey, 32)
	if err != nil {
		return nil, err
	}
	defer clear(ephemeralBytes)
	ephemeral, err := ecdh.X25519().NewPublicKey(ephemeralBytes)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	shared, err := opener.privateKey.ECDH(ephemeral)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	defer clear(shared)
	aead, key, err := deriveAEAD(shared, authenticated)
	if err != nil {
		return nil, err
	}
	defer clear(key)
	nonce, err := decodeExact(envelope.Nonce, aead.NonceSize())
	if err != nil {
		return nil, err
	}
	defer clear(nonce)
	ciphertext, err := base64.RawURLEncoding.DecodeString(envelope.Ciphertext)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	defer clear(ciphertext)
	plaintext, err := aead.Open(nil, nonce, ciphertext, authenticated)
	if err != nil || len(plaintext) == 0 || len(plaintext) > MaximumPlaintextBytes {
		clear(plaintext)
		return nil, ErrInvalidEnvelope
	}
	return plaintext, nil
}
