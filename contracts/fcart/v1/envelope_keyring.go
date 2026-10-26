package fcartv1

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type EnvelopeKeyring struct {
	current KEK
	keys    map[string]KEK
	random  io.Reader
}

func NewEnvelopeKeyring(
	current KEK,
	decryptOnly []KEK,
	random io.Reader,
) (*EnvelopeKeyring, error) {
	if current.id == "" || len(decryptOnly) > 7 {
		return nil, ErrInvalidKEK
	}
	if random == nil {
		random = rand.Reader
	}
	keys := map[string]KEK{current.id: current}
	for _, key := range decryptOnly {
		if key.id == "" {
			return nil, ErrInvalidKEK
		}
		if _, duplicate := keys[key.id]; duplicate {
			return nil, ErrInvalidKEK
		}
		keys[key.id] = key
	}
	return &EnvelopeKeyring{current: current, keys: keys, random: random}, nil
}

func (keyring *EnvelopeKeyring) Generate(
	ctx context.Context,
	identity KeyIdentity,
) (DEK, KeyEnvelope, error) {
	if keyring == nil || ctx == nil || identity.Validate() != nil {
		return DEK{}, KeyEnvelope{}, ErrInvalidKeyEnvelope
	}
	var key DEK
	if _, err := io.ReadFull(keyring.random, key[:]); err != nil {
		return DEK{}, KeyEnvelope{}, errors.Join(ErrRandomSource, err)
	}
	if err := key.Validate(); err != nil {
		clear(key[:])
		return DEK{}, KeyEnvelope{}, err
	}
	envelope, err := keyring.wrap(identity, keyring.current, key)
	if err != nil {
		clear(key[:])
		return DEK{}, KeyEnvelope{}, err
	}
	return key, envelope, nil
}

func (keyring *EnvelopeKeyring) Unwrap(
	ctx context.Context,
	identity KeyIdentity,
	envelope KeyEnvelope,
) (DEK, error) {
	if keyring == nil || ctx == nil || identity.Validate() != nil ||
		envelope.Validate() != nil {
		return DEK{}, ErrInvalidKeyEnvelope
	}
	key, found := keyring.keys[envelope.KEKID]
	if !found {
		return DEK{}, ErrKEKUnavailable
	}
	aead, err := keyAEAD(key)
	if err != nil {
		return DEK{}, err
	}
	aad, err := identity.aad(envelope.KEKID)
	if err != nil {
		return DEK{}, err
	}
	plaintext, err := aead.Open(nil, envelope.Nonce[:], envelope.Ciphertext, aad)
	if err != nil {
		return DEK{}, ErrKeyAuthentication
	}
	defer clear(plaintext)
	return NewDEK(plaintext)
}

func keyAEAD(key KEK) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key.material[:])
	if err != nil {
		return nil, ErrInvalidKEK
	}
	return cipher.NewGCM(block)
}
