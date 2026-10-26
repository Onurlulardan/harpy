package gcpkms

import (
	"context"
	"crypto/rand"
	"io"
	"sync"
	"sync/atomic"
)

type TOTPSealer struct {
	client     rawEncryptClient
	active     keyVersion
	randomness io.Reader
	close      sync.Once
	closeErr   error
	closed     atomic.Bool
}

func OpenTOTPSealer(ctx context.Context, activeKeyID string) (*TOTPSealer, error) {
	return openTOTPSealer(ctx, activeKeyID, openGoogleEncryptClient, rand.Reader)
}

func openTOTPSealer(
	ctx context.Context,
	activeKeyID string,
	opener encryptClientOpener,
	randomness io.Reader,
) (*TOTPSealer, error) {
	if ctx == nil || ctx.Err() != nil || opener == nil || randomness == nil {
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
	return &TOTPSealer{client: client, active: active, randomness: randomness}, nil
}

func (sealer *TOTPSealer) Close() error {
	if sealer == nil {
		return nil
	}
	sealer.close.Do(func() {
		sealer.closed.Store(true)
		if sealer.client != nil {
			sealer.closeErr = sealer.client.Close()
		}
	})
	return sealer.closeErr
}
