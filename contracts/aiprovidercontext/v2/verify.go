package aiprovidercontextv2

import (
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/base64"
)

type Verifier struct{ keys map[string]ed25519.PublicKey }

func NewVerifier(keys map[string][]byte) (*Verifier, error) {
	if len(keys) == 0 {
		return nil, ErrInvalidContext
	}
	owned := make(map[string]ed25519.PublicKey, len(keys))
	for keyID, key := range keys {
		if keyID == "" || len(key) != ed25519.PublicKeySize {
			return nil, ErrInvalidContext
		}
		owned[keyID] = append(ed25519.PublicKey(nil), key...)
	}
	return &Verifier{keys: owned}, nil
}

func (verifier *Verifier) Verify(document []byte, expected Expected) (Claims, error) {
	if verifier == nil || !expected.valid() {
		return Claims{}, ErrInvalidContext
	}
	token, err := Parse(document)
	if err != nil {
		return Claims{}, err
	}
	key, exists := verifier.keys[token.Claims.KeyID]
	if !exists || len(key) != ed25519.PublicKeySize {
		return Claims{}, ErrInvalidContext
	}
	payload, err := claimsBytes(token.Claims)
	if err != nil {
		return Claims{}, err
	}
	defer clear(payload)
	signature, err := base64.RawURLEncoding.DecodeString(token.Signature)
	if err != nil || !ed25519.Verify(key, payload, signature) {
		return Claims{}, ErrInvalidContext
	}
	now := expected.Now.UTC().Unix()
	if now < token.Claims.NotBefore || now >= token.Claims.ExpiresAt ||
		token.Claims.ExpiresAt > expected.HardDeadline.UTC().Unix() ||
		!same(token.Claims.TenantID, expected.TenantID) ||
		!same(token.Claims.ConfigID, expected.ConfigID) ||
		!same(token.Claims.Operation, expected.Operation) ||
		!same(token.Claims.InvocationID, expected.InvocationID) ||
		!same(token.Claims.ReceiptID, expected.ReceiptID) ||
		!same(token.Claims.RequestHash, expected.RequestHash) ||
		!same(token.Claims.CredentialEnvelopeHash, expected.CredentialEnvelopeHash) ||
		token.Claims.ConfigRevision != expected.ConfigRevision ||
		token.Claims.CredentialRevision != expected.CredentialRevision ||
		token.Claims.AuthorizationVersion != expected.AuthorizationVersion {
		return Claims{}, ErrInvalidContext
	}
	return token.Claims, nil
}

func (expected Expected) valid() bool {
	return !expected.Now.IsZero() && expected.HardDeadline.After(expected.Now) &&
		expected.ConfigRevision > 0 && expected.CredentialRevision >= 0 &&
		expected.AuthorizationVersion > 0
}

func same(actual, expected string) bool {
	return len(actual) == len(expected) &&
		subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}
