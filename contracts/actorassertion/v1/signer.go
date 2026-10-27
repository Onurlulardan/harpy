package actorassertionv1

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Signer struct {
	mu     sync.Mutex
	keyID  string
	key    ed25519.PrivateKey
	random io.Reader
	closed bool
}

func NewProductionSigner(keyID string, key ed25519.PrivateKey) (*Signer, error) {
	return NewSigner(keyID, key, rand.Reader)
}

func NewSigner(keyID string, key ed25519.PrivateKey, random io.Reader) (*Signer, error) {
	if !keyIDPattern.MatchString(keyID) || len(key) != ed25519.PrivateKeySize || random == nil {
		return nil, ErrContract
	}
	derived := ed25519.NewKeyFromSeed(key[:ed25519.SeedSize])
	if !derived.Equal(key) {
		return nil, ErrContract
	}
	return &Signer{keyID: keyID, key: append(ed25519.PrivateKey(nil), key...), random: random}, nil
}

func (signer *Signer) Issue(decision DecisionEvidence, now time.Time) (Credential, error) {
	if signer == nil || decision.Validate() != nil || now.IsZero() {
		return Credential{}, ErrContract
	}
	signer.mu.Lock()
	defer signer.mu.Unlock()
	if signer.closed || len(signer.key) != ed25519.PrivateKeySize {
		return Credential{}, ErrClosed
	}
	var randomID [16]byte
	if _, err := io.ReadFull(signer.random, randomID[:]); err != nil {
		return Credential{}, ErrContract
	}
	randomID[6] = (randomID[6] & 0x0f) | 0x40
	randomID[8] = (randomID[8] & 0x3f) | 0x80
	claims, err := newClaims(signer.keyID, decision, uuid.UUID(randomID), now.Unix())
	if err != nil {
		return Credential{}, err
	}
	payload, err := encodeClaims(claims)
	if err != nil {
		return Credential{}, err
	}
	signingInput := wirePrefix + "." + payload
	signature := ed25519.Sign(signer.key, []byte(signingInput))
	wire := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
	if len(wire) > MaximumWireSize {
		return Credential{}, ErrContract
	}
	return Credential{wire: wire}, nil
}

func (signer *Signer) Close() {
	if signer == nil {
		return
	}
	signer.mu.Lock()
	defer signer.mu.Unlock()
	for index := range signer.key {
		signer.key[index] = 0
	}
	signer.key = nil
	signer.closed = true
}
