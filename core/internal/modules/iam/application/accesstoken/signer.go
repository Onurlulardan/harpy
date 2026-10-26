package accesstoken

import (
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/json"
	"sync"
)

type Signer struct {
	mutex   sync.RWMutex
	profile Profile
	keyID   string
	key     ed25519.PrivateKey
}

func NewSigner(profile Profile, keyID string, key ed25519.PrivateKey) (*Signer, error) {
	if profile.issuer == "" || profile.audience == "" || !validKeyID(keyID) ||
		len(key) != ed25519.PrivateKeySize || !consistentPrivateKey(key) {
		return nil, ErrInvalidKey
	}
	return &Signer{
		profile: profile, keyID: keyID,
		key: append(ed25519.PrivateKey(nil), key...),
	}, nil
}

func (signer *Signer) Sign(claims Claims) (string, error) {
	if signer == nil || !claims.validShape() ||
		claims.issuer != signer.profile.issuer || claims.audience != signer.profile.audience {
		return "", ErrInvalidClaims
	}
	header, err := json.Marshal(headerDocument{Type: Type, Algorithm: Algorithm, KeyID: signer.keyID})
	if err != nil {
		return "", ErrInvalidToken
	}
	payload, err := json.Marshal(claims.document())
	if err != nil {
		return "", ErrInvalidToken
	}
	signingInput := encodeBase64URL(header) + "." + encodeBase64URL(payload)
	signer.mutex.RLock()
	defer signer.mutex.RUnlock()
	if len(signer.key) != ed25519.PrivateKeySize {
		return "", ErrSignerClosed
	}
	signature := ed25519.Sign(signer.key, []byte(signingInput))
	token := signingInput + "." + encodeBase64URL(signature)
	if len(token) > MaximumWireSize {
		return "", ErrInvalidClaims
	}
	return token, nil
}

func (signer *Signer) Close() {
	if signer == nil {
		return
	}
	signer.mutex.Lock()
	defer signer.mutex.Unlock()
	for index := range signer.key {
		signer.key[index] = 0
	}
	signer.key = nil
}

func consistentPrivateKey(key ed25519.PrivateKey) bool {
	derived := ed25519.NewKeyFromSeed(key[:ed25519.SeedSize])
	consistent := subtle.ConstantTimeCompare(derived, key) == 1
	for index := range derived {
		derived[index] = 0
	}
	return consistent
}
