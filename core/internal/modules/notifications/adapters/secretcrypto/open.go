package secretcrypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (sealer *Sealer) Open(ctx context.Context,
	secret ports.EncryptedSecret,
) ([]byte, error) {
	if sealer == nil || ctx == nil || !secret.ID.Valid() || secret.KEKRevision < 1 ||
		len(secret.Ciphertext) == 0 || len(secret.Nonce) != 12 || len(secret.WrappedDEK) == 0 {
		return nil, ErrSecretCrypto
	}
	dek, err := sealer.wrapper.Unwrap(ctx, secret.WrappedDEK, secret.KEKRevision)
	if err != nil || len(dek) != 32 {
		clear(dek)
		return nil, fmt.Errorf("unwrap notification DEK: %w", ErrSecretCrypto)
	}
	block, err := aes.NewCipher(dek)
	clear(dek)
	if err != nil {
		return nil, ErrSecretCrypto
	}
	aead, _ := cipher.NewGCM(block)
	plaintext, err := aead.Open(nil, secret.Nonce, secret.Ciphertext,
		secretAAD(secret.ID))
	if err != nil || sha256.Sum256(plaintext) != secret.Digest {
		clear(plaintext)
		return nil, fmt.Errorf("authenticate notification secret: %w", ErrSecretCrypto)
	}
	return plaintext, nil
}
