package gcpkms

import (
	"context"
	"sync"
	"sync/atomic"
)

type DeliveryOpener struct {
	client   rawDecryptClient
	close    sync.Once
	closeErr error
	closed   atomic.Bool
}

func OpenDeliveryOpener(ctx context.Context) (*DeliveryOpener, error) {
	return openDeliveryOpener(ctx, openGoogleClient)
}

func openDeliveryOpener(
	ctx context.Context,
	opener clientOpener,
) (*DeliveryOpener, error) {
	if ctx == nil || ctx.Err() != nil || opener == nil {
		return nil, ErrContract
	}
	client, err := opener(ctx)
	if err != nil || client == nil {
		return nil, ErrUnavailable
	}
	return &DeliveryOpener{client: client}, nil
}

func (opener *DeliveryOpener) Close() error {
	if opener == nil {
		return nil
	}
	opener.close.Do(func() {
		opener.closed.Store(true)
		if opener.client != nil {
			opener.closeErr = opener.client.Close()
		}
	})
	return opener.closeErr
}
