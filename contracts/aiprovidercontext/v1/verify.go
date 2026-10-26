package aiprovidercontextv1

import (
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/base64"
	"time"
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
	if verifier == nil || expected.Now.IsZero() || expected.ConfigRevision < 1 ||
		expected.AuthorizationVersion < 1 {
		return Claims{}, ErrInvalidContext
	}
	token, err := Parse(document)
	if err != nil {
		return Claims{}, err
	}
	key := verifier.keys[token.Claims.KeyID]
	payload, err := claimsBytes(token.Claims)
	if err != nil {
		return Claims{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(token.Signature)
	if err != nil || !ed25519.Verify(key, payload, signature) {
		return Claims{}, ErrInvalidContext
	}
	now := expected.Now.UTC().Unix()
	if now < token.Claims.NotBefore || now >= token.Claims.ExpiresAt ||
		subtle.ConstantTimeCompare([]byte(token.Claims.TenantID), []byte(expected.TenantID)) != 1 ||
		subtle.ConstantTimeCompare([]byte(token.Claims.ConfigID), []byte(expected.ConfigID)) != 1 ||
		subtle.ConstantTimeCompare([]byte(token.Claims.Operation), []byte(expected.Operation)) != 1 ||
		token.Claims.ConfigRevision != expected.ConfigRevision ||
		token.Claims.AuthorizationVersion != expected.AuthorizationVersion {
		return Claims{}, ErrInvalidContext
	}
	return token.Claims, nil
}

func NewClaims(now time.Time, lifetime time.Duration) Claims {
	issued := now.UTC().Truncate(time.Second)
	return Claims{IssuedAt: issued.Unix(), NotBefore: issued.Unix(),
		ExpiresAt: issued.Add(lifetime).Unix()}
}
