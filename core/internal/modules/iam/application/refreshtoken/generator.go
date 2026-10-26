package refreshtoken

import (
	"crypto/rand"
	"io"
)

const (
	SelectorBytes = 16
	SecretBytes   = 32
)

type Generator struct{ randomness io.Reader }

func NewGenerator() Generator { return Generator{randomness: rand.Reader} }

func (generator Generator) Generate() (*Issued, error) {
	if generator.randomness == nil {
		return nil, ErrRandomUnavailable
	}
	token := &Issued{active: true}
	if _, err := io.ReadFull(generator.randomness, token.selector[:]); err != nil {
		token.Destroy()
		return nil, ErrRandomUnavailable
	}
	if _, err := io.ReadFull(generator.randomness, token.secret[:]); err != nil {
		token.Destroy()
		return nil, ErrRandomUnavailable
	}
	if token.selector == [SelectorBytes]byte{} || token.secret == [SecretBytes]byte{} {
		token.Destroy()
		return nil, ErrRandomUnavailable
	}
	return token, nil
}
