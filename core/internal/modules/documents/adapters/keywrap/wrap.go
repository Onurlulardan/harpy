package keywrap

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

func Wrap(dek, kek []byte, context Context, randomness io.Reader) ([]byte, error) {
	if len(dek) != 32 || len(kek) != 32 || randomness == nil {
		return nil, ErrInvalidWrap
	}
	header, err := header(context)
	if err != nil {
		return nil, err
	}
	aead, err := newGCM(kek)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(randomness, nonce); err != nil {
		return nil, err
	}
	result := make([]byte, 0, len(header)+len(nonce)+len(dek)+aead.Overhead())
	result = append(result, header...)
	result = append(result, nonce...)
	result = aead.Seal(result, nonce, dek, header)
	return result, nil
}

func Unwrap(value, kek []byte, expected Context) ([]byte, error) {
	if len(kek) != 32 || !expected.valid() {
		return nil, ErrInvalidWrap
	}
	actual, header, offset, err := decodeHeader(value)
	if err != nil || actual != expected {
		return nil, ErrInvalidWrap
	}
	aead, err := newGCM(kek)
	if err != nil {
		return nil, err
	}
	nonce := value[offset : offset+aead.NonceSize()]
	opened, err := aead.Open(nil, nonce, value[offset+aead.NonceSize():], header)
	if err != nil || len(opened) != 32 {
		return nil, ErrKeyTampered
	}
	return opened, nil
}

func Rewrap(value, oldKEK, newKEK []byte, oldContext, newContext Context,
	randomness io.Reader,
) ([]byte, error) {
	dek, err := Unwrap(value, oldKEK, oldContext)
	if err != nil {
		return nil, err
	}
	defer clear(dek)
	return Wrap(dek, newKEK, newContext, randomness)
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrInvalidWrap
	}
	result, err := cipher.NewGCM(block)
	if err != nil || result.NonceSize() != 12 {
		return nil, ErrInvalidWrap
	}
	return result, nil
}
