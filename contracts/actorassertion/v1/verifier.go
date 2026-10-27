package actorassertionv1

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"time"
)

const futureClockSkewSeconds = int64(2)

type Verifier struct {
	keys map[string]ed25519.PublicKey
}

func NewVerifier(keys map[string]ed25519.PublicKey) (Verifier, error) {
	if len(keys) < 1 || len(keys) > 8 {
		return Verifier{}, ErrContract
	}
	copied := make(map[string]ed25519.PublicKey, len(keys))
	for keyID, key := range keys {
		if !keyIDPattern.MatchString(keyID) || len(key) != ed25519.PublicKeySize {
			return Verifier{}, ErrContract
		}
		copied[keyID] = append(ed25519.PublicKey(nil), key...)
	}
	return Verifier{keys: copied}, nil
}

func (verifier Verifier) Verify(
	wire string,
	expectation Expectation,
	now time.Time,
) (Claims, error) {
	if len(wire) < 1 || len(wire) > MaximumWireSize || expectation.Validate() != nil || now.IsZero() {
		return Claims{}, ErrContract
	}
	parts := strings.Split(wire, ".")
	if len(parts) != 3 || parts[0] != wirePrefix {
		return Claims{}, ErrContract
	}
	claims, err := decodeClaims(parts[1])
	if err != nil {
		return Claims{}, err
	}
	key, exists := verifier.keys[claims.KeyID]
	if !exists {
		return Claims{}, ErrSignature
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(signature) != ed25519.SignatureSize ||
		base64.RawURLEncoding.EncodeToString(signature) != parts[2] ||
		!ed25519.Verify(key, []byte(parts[0]+"."+parts[1]), signature) {
		return Claims{}, ErrSignature
	}
	if !expectation.matches(claims) {
		return Claims{}, ErrContract
	}
	nowUnix := now.Unix()
	if nowUnix >= claims.ExpiresAt {
		return Claims{}, ErrExpired
	}
	if nowUnix+futureClockSkewSeconds < claims.NotBefore {
		return Claims{}, ErrContract
	}
	return claims, nil
}
