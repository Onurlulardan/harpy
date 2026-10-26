package gcpkms

import (
	"context"
	"crypto/rand"
	"io"
	"sync"
	"sync/atomic"
)

type DeliverySealer struct {
	client     rawEncryptClient
	active     keyVersion
	randomness io.Reader
	close      sync.Once
	closeErr   error
	closed     atomic.Bool
}

func OpenDeliverySealer(
	ctx context.Context,
	activeKeyID string,
) (*DeliverySealer, error) {
	return openDeliverySealer(ctx, activeKeyID, openGoogleEncryptClient, rand.Reader)
}

func openDeliverySealer(
	ctx context.Context,
	activeKeyID string,
	opener encryptClientOpener,
	randomness io.Reader,
) (*DeliverySealer, error) {
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
	return &DeliverySealer{
		client: client, active: active, randomness: randomness,
	}, nil
}

func (sealer *DeliverySealer) Close() error {
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
