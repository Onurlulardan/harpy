package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (binder *Binder) BindRealtime(ctx context.Context,
	scope api.Scope,
) (ports.RealtimeTicketStore, error) {
	stores, err := binder.Bind(ctx, scope)
	if err != nil {
		return nil, err
	}
	store, ok := stores.(*Store)
	if !ok || store == nil {
		return nil, ports.ErrBinding
	}
	return store, nil
}

func (binder *Binder) BindLegacyConnection(ctx context.Context,
	scope api.Scope,
) (ports.LegacyConnectionStore, error) {
	store, err := binder.BindRealtime(ctx, scope)
	if err != nil {
		return nil, err
	}
	legacy, ok := store.(ports.LegacyConnectionStore)
	if !ok || legacy == nil {
		return nil, ports.ErrBinding
	}
	return legacy, nil
}

func (binder *Binder) BindRealtimeProjection(ctx context.Context,
	scope api.Scope,
) (ports.RealtimeProjectionStore, error) {
	stores, err := binder.Bind(ctx, scope)
	if err != nil {
		return nil, err
	}
	projection, ok := stores.(ports.RealtimeProjectionStore)
	if !ok || projection == nil {
		return nil, ports.ErrBinding
	}
	return projection, nil
}

var _ ports.RealtimeTicketBinder = (*Binder)(nil)
var _ ports.LegacyConnectionBinder = (*Binder)(nil)
var _ ports.RealtimeProjectionBinder = (*Binder)(nil)
