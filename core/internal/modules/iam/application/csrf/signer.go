package csrf

import (
	"crypto/hmac"
	"crypto/rand"
	"io"
	"strconv"
	"sync"
	"time"
)

type Signer struct {
	mutex      sync.Mutex
	policy     OriginPolicy
	keyID      string
	key        []byte
	randomness io.Reader
}

func NewSigner(
	policy OriginPolicy,
	keyID string,
	key []byte,
	randomness io.Reader,
) (*Signer, error) {
	if len(policy.allowed) == 0 || !validKeyID(keyID) || len(key) != keyBytes || randomness == nil {
		return nil, ErrInvalidKey
	}
	return &Signer{
		policy: policy, keyID: keyID, key: append([]byte(nil), key...), randomness: randomness,
	}, nil
}

func NewProductionSigner(
	policy OriginPolicy,
	keyID string,
	key []byte,
) (*Signer, error) {
	return NewSigner(policy, keyID, key, rand.Reader)
}

func (signer *Signer) Sign(
	selector [16]byte,
	origin string,
	now time.Time,
) (*Credential, error) {
	if signer == nil || selector == [16]byte{} || !signer.policy.Allows(origin) ||
		now.IsZero() || now.Nanosecond() != 0 {
		return nil, ErrInvalidCredential
	}
	signer.mutex.Lock()
	defer signer.mutex.Unlock()
	if len(signer.key) != keyBytes {
		return nil, ErrSignerClosed
	}
	nonce := make([]byte, nonceBytes)
	if _, err := io.ReadFull(signer.randomness, nonce); err != nil || allZero(nonce) {
		wipe(nonce)
		return nil, ErrInvalidCredential
	}
	nonceText := encode(nonce)
	wipe(nonce)
	selectorText := encode(selector[:])
	issuedAt := strconv.FormatInt(now.Unix(), 10)
	expiry := strconv.FormatInt(now.Add(Lifetime).Unix(), 10)
	mac := authenticationCode(
		signer.key, signer.keyID, nonceText, selectorText, origin, issuedAt, expiry,
	)
	wire := []byte(
		signer.keyID + "." + nonceText + "." + issuedAt + "." + expiry + "." + encode(mac[:]),
	)
	wipe(mac[:])
	if len(wire) > MaximumWireSize {
		wipe(wire)
		return nil, ErrInvalidCredential
	}
	return &Credential{wire: wire, active: true}, nil
}

func (signer *Signer) Close() {
	if signer == nil {
		return
	}
	signer.mutex.Lock()
	defer signer.mutex.Unlock()
	wipe(signer.key)
	signer.key = nil
	signer.randomness = nil
}

func allZero(value []byte) bool {
	zero := make([]byte, len(value))
	result := hmac.Equal(value, zero)
	wipe(zero)
	return result
}
