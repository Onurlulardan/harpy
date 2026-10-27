package secretcrypto

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (sealer *Sealer) Rewrap(ctx context.Context, secret ports.EncryptedSecret,
	newRevision int64,
) (ports.EncryptedSecret, error) {
	if sealer == nil || ctx == nil || newRevision < 1 ||
		newRevision == secret.KEKRevision || !secret.ID.Valid() {
		return ports.EncryptedSecret{}, ErrSecretCrypto
	}
	dek, err := sealer.wrapper.Unwrap(ctx, secret.WrappedDEK, secret.KEKRevision)
	if err != nil || len(dek) != 32 {
		clear(dek)
		return ports.EncryptedSecret{}, ErrSecretCrypto
	}
	wrapped, err := sealer.wrapper.Wrap(ctx, dek, newRevision)
	clear(dek)
	if err != nil || len(wrapped) == 0 {
		return ports.EncryptedSecret{}, ErrSecretCrypto
	}
	secret.Ciphertext = append([]byte(nil), secret.Ciphertext...)
	secret.Nonce = append([]byte(nil), secret.Nonce...)
	secret.WrappedDEK = wrapped
	secret.KEKRevision = newRevision
	return secret, nil
}
