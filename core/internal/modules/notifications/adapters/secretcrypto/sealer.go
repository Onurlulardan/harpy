package secretcrypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"io"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

var ErrSecretCrypto = errors.New("notification secret cryptography failed")

type KeyWrapper interface {
	Wrap(context.Context, []byte, int64) ([]byte, error)
	Unwrap(context.Context, []byte, int64) ([]byte, error)
}

type Sealer struct {
	random  io.Reader
	wrapper KeyWrapper
}

func New(random io.Reader, wrapper KeyWrapper) (*Sealer, error) {
	if random == nil || nilInterface(wrapper) {
		return nil, ErrSecretCrypto
	}
	return &Sealer{random: random, wrapper: wrapper}, nil
}

func (sealer *Sealer) Seal(ctx context.Context, id domain.ID, plaintext []byte,
	kekRevision int64,
) (ports.EncryptedSecret, error) {
	if sealer == nil || ctx == nil || !id.Valid() || len(plaintext) == 0 ||
		len(plaintext) > 16*1024 || kekRevision < 1 {
		return ports.EncryptedSecret{}, ErrSecretCrypto
	}
	dek := make([]byte, 32)
	if _, err := io.ReadFull(sealer.random, dek); err != nil {
		return ports.EncryptedSecret{}, ErrSecretCrypto
	}
	block, err := aes.NewCipher(dek)
	if err != nil {
		clear(dek)
		return ports.EncryptedSecret{}, ErrSecretCrypto
	}
	aead, _ := cipher.NewGCM(block)
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(sealer.random, nonce); err != nil {
		clear(dek)
		return ports.EncryptedSecret{}, ErrSecretCrypto
	}
	ciphertext := aead.Seal(nil, nonce, plaintext, secretAAD(id))
	wrapped, err := sealer.wrapper.Wrap(ctx, dek, kekRevision)
	clear(dek)
	if err != nil || len(wrapped) == 0 {
		return ports.EncryptedSecret{}, ErrSecretCrypto
	}
	return ports.EncryptedSecret{ID: id, Ciphertext: ciphertext, Nonce: nonce,
		WrappedDEK: wrapped, KEKRevision: kekRevision,
		Digest: sha256.Sum256(plaintext)}, nil
}

func secretAAD(id domain.ID) []byte {
	const prefix = "FCNOTIFYSECRET1"
	result := make([]byte, len(prefix)+len(id))
	copy(result, prefix)
	copy(result[len(prefix):], id[:])
	return result
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}

var _ ports.SecretSealer = (*Sealer)(nil)
var _ ports.SecretOpener = (*Sealer)(nil)
