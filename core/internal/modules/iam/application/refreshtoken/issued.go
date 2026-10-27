package refreshtoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

const wireBytes = 22 + 1 + 43

type Issued struct {
	selector [SelectorBytes]byte
	secret   [SecretBytes]byte
	active   bool
}

func (token *Issued) Selector() ([SelectorBytes]byte, error) {
	if token == nil || !token.active {
		return [SelectorBytes]byte{}, ErrTokenDestroyed
	}
	return token.selector, nil
}

func (token *Issued) SecretHMAC(key []byte) ([32]byte, error) {
	if token == nil || !token.active {
		return [32]byte{}, ErrTokenDestroyed
	}
	return secretHMAC(key, token.secret[:])
}

// RevealOnce gives the canonical token to the cookie adapter and wipes both
// the temporary wire buffer and this token before returning.
func (token *Issued) RevealOnce(consume func([]byte) error) error {
	if token == nil || !token.active {
		return ErrTokenDestroyed
	}
	if consume == nil {
		return ErrTokenInvalid
	}
	wire := make([]byte, wireBytes)
	base64.RawURLEncoding.Encode(wire[:22], token.selector[:])
	wire[22] = '.'
	base64.RawURLEncoding.Encode(wire[23:], token.secret[:])
	err := consume(wire)
	wipe(wire)
	token.Destroy()
	return err
}

func (token *Issued) Destroy() {
	if token == nil {
		return
	}
	wipe(token.selector[:])
	wipe(token.secret[:])
	token.active = false
}

func secretHMAC(key []byte, secret []byte) ([32]byte, error) {
	if len(key) < 32 {
		return [32]byte{}, ErrKeyInvalid
	}
	hash := hmac.New(sha256.New, key)
	_, _ = hash.Write(secret)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func wipe(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
