package secretcrypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"io"
)

type AESKeyWrapper struct {
	keys   map[int64][32]byte
	random io.Reader
}

func NewAESKeyWrapper(keys map[int64][32]byte, random io.Reader) (*AESKeyWrapper, error) {
	if len(keys) == 0 || random == nil {
		return nil, ErrSecretCrypto
	}
	copyKeys := make(map[int64][32]byte, len(keys))
	for revision, key := range keys {
		if revision < 1 || key == [32]byte{} {
			return nil, ErrSecretCrypto
		}
		copyKeys[revision] = key
	}
	return &AESKeyWrapper{keys: copyKeys, random: random}, nil
}

func (wrapper *AESKeyWrapper) Wrap(ctx context.Context, plaintext []byte,
	revision int64,
) ([]byte, error) {
	if wrapper == nil || ctx == nil || len(plaintext) != 32 {
		return nil, ErrSecretCrypto
	}
	aead, err := wrapper.aead(revision)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(wrapper.random, nonce); err != nil {
		return nil, ErrSecretCrypto
	}
	return append(nonce, aead.Seal(nil, nonce, plaintext, keyAAD(revision))...), nil
}

func (wrapper *AESKeyWrapper) Unwrap(ctx context.Context, wrapped []byte,
	revision int64,
) ([]byte, error) {
	if wrapper == nil || ctx == nil {
		return nil, ErrSecretCrypto
	}
	aead, err := wrapper.aead(revision)
	if err != nil || len(wrapped) <= aead.NonceSize() {
		return nil, ErrSecretCrypto
	}
	result, err := aead.Open(nil, wrapped[:aead.NonceSize()],
		wrapped[aead.NonceSize():], keyAAD(revision))
	if err != nil || len(result) != 32 {
		clear(result)
		return nil, ErrSecretCrypto
	}
	return result, nil
}

func (wrapper *AESKeyWrapper) aead(revision int64) (cipher.AEAD, error) {
	key, found := wrapper.keys[revision]
	if !found {
		return nil, ErrSecretCrypto
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, ErrSecretCrypto
	}
	return cipher.NewGCM(block)
}

func keyAAD(revision int64) []byte {
	result := make([]byte, 24)
	copy(result, "FCNOTIFYWRAP1")
	binary.BigEndian.PutUint64(result[16:], uint64(revision))
	return result
}
