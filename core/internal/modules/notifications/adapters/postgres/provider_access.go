package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) SelectProvider(ctx context.Context, channel domain.Channel,
	requested domain.ID,
) (ports.ProviderConfig, error) {
	if !store.usable(ctx, false) ||
		(channel != domain.ChannelEmail && channel != domain.ChannelTelegram) {
		return ports.ProviderConfig{}, ErrInvalid
	}
	requestedText := ""
	if requested.Valid() {
		requestedText = requested.String()
	}
	value, err := scanProvider(store.queryRow(ctx, resolveProviderSQL,
		string(channel), requestedText))
	if store.notFound(err) {
		return ports.ProviderConfig{}, ports.ErrProviderNotFound
	}
	return value, err
}

func (store *Store) LoadProviderConfig(ctx context.Context,
	id domain.ID,
) (ports.ProviderConfig, error) {
	if !store.usable(ctx, false) || !id.Valid() {
		return ports.ProviderConfig{}, ErrInvalid
	}
	value, err := scanProvider(store.queryRow(ctx, getProviderConfigSQL, id.String()))
	if store.notFound(err) {
		return ports.ProviderConfig{}, ports.ErrProviderNotFound
	}
	return value, err
}
