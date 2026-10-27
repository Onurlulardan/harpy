package gcpkms

import (
	"context"
	"sync"
	"sync/atomic"
)

type TOTPDecryptor struct {
	client   rawDecryptClient
	active   keyVersion
	close    sync.Once
	closeErr error
	closed   atomic.Bool
}

func OpenTOTPDecryptor(
	ctx context.Context,
	activeKeyID string,
) (*TOTPDecryptor, error) {
	return openTOTPDecryptor(ctx, activeKeyID, openGoogleClient)
}

func openTOTPDecryptor(
	ctx context.Context,
	activeKeyID string,
	opener clientOpener,
) (*TOTPDecryptor, error) {
	if ctx == nil || ctx.Err() != nil || opener == nil {
		return nil, ErrContract
	}
	active, err := parseKeyVersion(activeKeyID)
	if err != nil {
		return nil, err
	}
	client, err := opener(ctx)
	if err != nil || client == nil {
		return nil, ErrUnavailable
	}
	return &TOTPDecryptor{client: client, active: active}, nil
}

func (decryptor *TOTPDecryptor) Close() error {
	if decryptor == nil {
		return nil
	}
	decryptor.close.Do(func() {
		decryptor.closed.Store(true)
		if decryptor.client != nil {
			decryptor.closeErr = decryptor.client.Close()
		}
	})
	return decryptor.closeErr
}
