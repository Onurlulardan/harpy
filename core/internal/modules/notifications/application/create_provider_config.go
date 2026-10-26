package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type CreateProviderConfigInput struct {
	Kind        ports.ProviderKind
	Name        string
	Enabled     bool
	Metadata    ports.ProviderMetadata
	Secret      []byte
	KEKRevision int64
}

func (application *ProviderApplication) CreateProviderConfig(ctx context.Context,
	scope api.Scope, input CreateProviderConfigInput,
) (ports.ProviderConfig, error) {
	defer clear(input.Secret)
	input.Name = strings.TrimSpace(input.Name)
	if application == nil || input.Name == "" || len(input.Name) > 128 ||
		!validateProviderMetadata(input.Kind, input.Metadata) ||
		len(input.Secret) == 0 || len(input.Secret) > 16*1024 || input.KEKRevision < 1 {
		return ports.ProviderConfig{}, ErrInvalid
	}
	stores, err := application.base.bindOperation(ctx, scope, PermissionSend)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	now := application.base.clock.Now()
	configID, err := application.base.nextID(now)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	versionID, err := application.base.nextID(now)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	secretID, err := application.base.nextID(now)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	sealed, err := application.sealer.Seal(ctx, secretID, input.Secret, input.KEKRevision)
	if err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	config := ports.ProviderConfig{ID: configID, VersionID: versionID, SecretID: secretID,
		Kind: input.Kind, Name: input.Name, Enabled: input.Enabled, Revision: 1,
		Metadata: input.Metadata, MetadataDigest: providerMetadataDigest(input.Kind, input.Metadata)}
	if err = stores.Providers().StoreProviderSecret(ctx, sealed); err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	if err = stores.Providers().CreateProviderConfig(ctx, config); err != nil {
		return ports.ProviderConfig{}, ErrConflict
	}
	if err = stores.Evidence().AppendAudit(ctx, "notifications.provider.create",
		configID, scopeActor(scope), config.Revision); err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	return config, nil
}
