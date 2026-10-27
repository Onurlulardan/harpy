package webhookbody

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/subtle"
)

type ReconciliationOpener struct {
	unwrapper Unwrapper
}

func NewReconciliationOpener(unwrapper Unwrapper) (*ReconciliationOpener, error) {
	if unwrapper == nil {
		return nil, ErrConfiguration
	}
	return &ReconciliationOpener{unwrapper: unwrapper}, nil
}

func (opener *ReconciliationOpener) Open(
	ctx context.Context,
	associatedData []byte,
	envelope Envelope,
) ([]byte, error) {
	if len(associatedData) == 0 || !envelope.valid() {
		return nil, ErrEnvelope
	}
	dek, err := opener.unwrapper.Unwrap(ctx, envelope.WrappedDEK)
	if err != nil || len(dek) != dekSize {
		clear(dek)
		return nil, ErrCryptography
	}
	defer clear(dek)
	block, err := aes.NewCipher(dek)
	if err != nil {
		return nil, ErrCryptography
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrCryptography
	}
	body, err := gcm.Open(nil, envelope.Nonce, envelope.Ciphertext, associatedData)
	if err != nil || len(body) == 0 || len(body) > MaximumBodySize {
		return nil, ErrCryptography
	}
	digest := sha256.Sum256(body)
	if subtle.ConstantTimeCompare(digest[:], envelope.BodySHA256[:]) != 1 {
		clear(body)
		return nil, ErrCryptography
	}
	return body, nil
}
