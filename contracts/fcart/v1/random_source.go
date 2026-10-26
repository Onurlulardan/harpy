package fcartv1

import (
	"crypto/rand"
	"io"
)

type RandomSource interface {
	Nonce() ([NonceBytes]byte, error)
	isFCART1RandomSource()
}

type cryptographicRandomSource struct{}

func NewCryptographicRandomSource() RandomSource {
	return cryptographicRandomSource{}
}

func (cryptographicRandomSource) Nonce() ([NonceBytes]byte, error) {
	var nonce [NonceBytes]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nonce, ErrRandomSource
	}
	return nonce, nil
}

func (cryptographicRandomSource) isFCART1RandomSource() {}
