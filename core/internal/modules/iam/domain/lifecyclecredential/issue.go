package lifecyclecredential

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

type Generator struct{ randomness io.Reader }

func NewGenerator() Generator {
	return Generator{randomness: rand.Reader}
}

type IssuedCredential struct {
	selector [16]byte
	secret   [32]byte
	active   bool
}

func (generator Generator) Generate(
	tenantID ID,
	purpose Purpose,
	tokenID ID,
	keys OneTimeKeyring,
) (*IssuedCredential, TokenIssue, error) {
	if tenantID == [16]byte{} || tokenID == [16]byte{} || !purpose.Valid() ||
		keys == nil || generator.randomness == nil {
		return nil, TokenIssue{}, ErrInvalid
	}
	keyID, key, err := keys.Active()
	if err != nil || keyID == "" || key == [32]byte{} {
		clear(key[:])
		return nil, TokenIssue{}, ErrUnavailable
	}
	value := &IssuedCredential{active: true}
	_, selectorErr := io.ReadFull(generator.randomness, value.selector[:])
	_, secretErr := io.ReadFull(generator.randomness, value.secret[:])
	if selectorErr != nil || secretErr != nil || value.selector == [16]byte{} ||
		value.secret == [32]byte{} {
		value.Destroy()
		clear(key[:])
		return nil, TokenIssue{}, ErrUnavailable
	}
	digest, digestErr := Digest(key, tenantID, purpose, value.selector, value.secret)
	clear(key[:])
	binding, bindingErr := BindingSHA256(tenantID, purpose, value.selector)
	if digestErr != nil || bindingErr != nil {
		value.Destroy()
		return nil, TokenIssue{}, ErrInvalid
	}
	return value, TokenIssue{ID: tokenID, Selector: value.selector, HashKeyID: keyID,
		SecretHMAC: digest, BindingSHA256: binding}, nil
}

func (value *IssuedCredential) RevealOnce(consume func([]byte) error) error {
	if value == nil || !value.active || consume == nil {
		return ErrInvalid
	}
	wire := make([]byte, WireBytes)
	base64.RawURLEncoding.Encode(wire[:22], value.selector[:])
	wire[22] = '.'
	base64.RawURLEncoding.Encode(wire[23:], value.secret[:])
	err := consume(wire)
	clear(wire)
	value.Destroy()
	return err
}

func (value *IssuedCredential) Destroy() {
	if value == nil {
		return
	}
	clear(value.selector[:])
	clear(value.secret[:])
	value.active = false
}
