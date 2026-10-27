package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (value *testStores) SelectProvider(context.Context, domain.Channel,
	domain.ID,
) (ports.ProviderConfig, error) {
	if !value.provider.VersionID.Valid() {
		return ports.ProviderConfig{}, ports.ErrProviderNotFound
	}
	return value.provider, nil
}

func (value *testStores) LoadProviderConfig(_ context.Context,
	id domain.ID,
) (ports.ProviderConfig, error) {
	if value.provider.ID != id {
		return ports.ProviderConfig{}, ports.ErrProviderNotFound
	}
	return value.provider, nil
}

func (value *testStores) CreateProviderConfig(_ context.Context,
	config ports.ProviderConfig,
) error {
	value.provider = config
	return nil
}

func (value *testStores) UpdateProviderConfig(_ context.Context,
	config ports.ProviderConfig, expected int64,
) error {
	if value.provider.Revision != expected {
		return ErrConflict
	}
	value.provider = config
	return nil
}

func (value *testStores) LoadProviderSecret(_ context.Context,
	id domain.ID,
) (ports.EncryptedSecret, error) {
	secret, found := value.providerSecrets[id]
	if !found {
		return ports.EncryptedSecret{}, ports.ErrProviderNotFound
	}
	return secret, nil
}

func (value *testStores) StoreProviderSecret(_ context.Context,
	secret ports.EncryptedSecret,
) error {
	value.providerSecrets[secret.ID] = secret
	return nil
}

func (value *testStores) RotateProviderSecret(_ context.Context,
	config ports.ProviderConfig, secret ports.EncryptedSecret, expected int64,
) error {
	if value.provider.Revision != expected {
		return ErrConflict
	}
	value.provider = config
	value.providerSecrets[secret.ID] = secret
	return nil
}

func (value *testStores) RewrapProviderSecret(_ context.Context,
	secret ports.EncryptedSecret, expected int64,
) error {
	current, found := value.providerSecrets[secret.ID]
	if !found || current.KEKRevision != expected {
		return ErrConflict
	}
	value.providerSecrets[secret.ID] = secret
	return nil
}
