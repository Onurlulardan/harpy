package refreshtoken

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
)

type Presented struct {
	selector [SelectorBytes]byte
	secret   [SecretBytes]byte
	active   bool
}

func Parse(wire []byte) (*Presented, error) {
	if len(wire) != wireBytes || wire[22] != '.' {
		return nil, ErrTokenInvalid
	}
	token := &Presented{active: true}
	if _, err := base64.RawURLEncoding.Decode(token.selector[:], wire[:22]); err != nil {
		token.Destroy()
		return nil, ErrTokenInvalid
	}
	if _, err := base64.RawURLEncoding.Decode(token.secret[:], wire[23:]); err != nil {
		token.Destroy()
		return nil, ErrTokenInvalid
	}
	canonical := make([]byte, wireBytes)
	base64.RawURLEncoding.Encode(canonical[:22], token.selector[:])
	canonical[22] = '.'
	base64.RawURLEncoding.Encode(canonical[23:], token.secret[:])
	valid := bytes.Equal(canonical, wire) && token.selector != [SelectorBytes]byte{} &&
		token.secret != [SecretBytes]byte{}
	wipe(canonical)
	if !valid {
		token.Destroy()
		return nil, ErrTokenInvalid
	}
	return token, nil
}

func (token *Presented) Selector() ([SelectorBytes]byte, error) {
	if token == nil || !token.active {
		return [SelectorBytes]byte{}, ErrTokenDestroyed
	}
	return token.selector, nil
}

func (token *Presented) VerifyHMAC(key []byte, expected [32]byte) (bool, error) {
	if token == nil || !token.active {
		return false, ErrTokenDestroyed
	}
	actual, err := secretHMAC(key, token.secret[:])
	if err != nil {
		return false, err
	}
	matched := subtle.ConstantTimeCompare(actual[:], expected[:]) == 1
	wipe(actual[:])
	return matched, nil
}

func (token *Presented) Destroy() {
	if token == nil {
		return
	}
	wipe(token.selector[:])
	wipe(token.secret[:])
	token.active = false
}
