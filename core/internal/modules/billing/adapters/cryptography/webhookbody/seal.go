package webhookbody

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"io"
)

type Sealer struct {
	random  io.Reader
	wrapper Wrapper
}

func NewSealer(random io.Reader, wrapper Wrapper) (*Sealer, error) {
	if random == nil || wrapper == nil {
		return nil, ErrConfiguration
	}
	return &Sealer{random: random, wrapper: wrapper}, nil
}

func (sealer *Sealer) Seal(
	ctx context.Context,
	associatedData []byte,
	body []byte,
) (Envelope, error) {
	if len(associatedData) == 0 || len(body) == 0 || len(body) > MaximumBodySize {
		return Envelope{}, ErrBody
	}
	dek := make([]byte, dekSize)
	defer clear(dek)
	if _, err := io.ReadFull(sealer.random, dek); err != nil {
		return Envelope{}, ErrCryptography
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		return Envelope{}, ErrCryptography
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Envelope{}, ErrCryptography
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(sealer.random, nonce); err != nil {
		return Envelope{}, ErrCryptography
	}
	wrapped, err := sealer.wrapper.Wrap(ctx, dek)
	if err != nil || wrapped.KeyID == "" || len(wrapped.Ciphertext) == 0 {
		return Envelope{}, ErrCryptography
	}
	return Envelope{
		Algorithm:  Algorithm,
		Nonce:      append([]byte(nil), nonce...),
		Ciphertext: gcm.Seal(nil, nonce, body, associatedData),
		WrappedDEK: WrappedDEK{
			KeyID: wrapped.KeyID, Ciphertext: append([]byte(nil), wrapped.Ciphertext...),
		},
		BodySHA256: sha256.Sum256(body),
	}, nil
}
