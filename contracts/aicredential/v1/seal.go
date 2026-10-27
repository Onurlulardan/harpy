package aicredentialv1

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const derivationInfo = "harpy:ai-credential-envelope:v1"

type Sealer struct{ publicKey *ecdh.PublicKey }

func NewSealer(publicKey []byte) (*Sealer, error) {
	parsed, err := ecdh.X25519().NewPublicKey(publicKey)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	return &Sealer{publicKey: parsed}, nil
}

func (sealer *Sealer) Seal(plaintext []byte, aad AAD) ([]byte, error) {
	if sealer == nil || sealer.publicKey == nil || len(plaintext) == 0 ||
		len(plaintext) > MaximumPlaintextBytes {
		return nil, ErrInvalidEnvelope
	}
	authenticated, err := aadBytes(aad)
	if err != nil {
		return nil, err
	}
	ephemeral, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	shared, err := ephemeral.ECDH(sealer.publicKey)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	defer clear(shared)
	aead, key, err := deriveAEAD(shared, authenticated)
	if err != nil {
		return nil, err
	}
	defer clear(key)
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, ErrInvalidEnvelope
	}
	ciphertext := aead.Seal(nil, nonce, plaintext, authenticated)
	envelope := Envelope{ContractVersion: ContractVersion, AAD: aad,
		EphemeralPublicKey: base64.RawURLEncoding.EncodeToString(ephemeral.PublicKey().Bytes()),
		Nonce:              base64.RawURLEncoding.EncodeToString(nonce),
		Ciphertext:         base64.RawURLEncoding.EncodeToString(ciphertext)}
	document, err := json.Marshal(envelope)
	clear(ciphertext)
	clear(nonce)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumEnvelopeBytes)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	return canonical, nil
}

func deriveAEAD(shared, authenticated []byte) (cipher.AEAD, []byte, error) {
	salt := sha256.Sum256(authenticated)
	key, err := hkdf.Key(sha256.New, shared, salt[:], derivationInfo, 32)
	if err != nil {
		return nil, nil, ErrInvalidEnvelope
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		clear(key)
		return nil, nil, ErrInvalidEnvelope
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		clear(key)
		return nil, nil, ErrInvalidEnvelope
	}
	return aead, key, nil
}
