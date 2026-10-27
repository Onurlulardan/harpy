package application

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type UpdateProviderConfigInput struct {
	ID               domain.ID
	Name             string
	Enabled          bool
	Metadata         ports.ProviderMetadata
	ExpectedRevision int64
}

func (application *ProviderApplication) UpdateProviderConfig(ctx context.Context,
	scope api.Scope, input UpdateProviderConfigInput,
) (ports.ProviderConfig, error) {
	input.Name = strings.TrimSpace(input.Name)
	if application == nil || !input.ID.Valid() || input.Name == "" ||
		len(input.Name) > 128 || input.ExpectedRevision < 1 {
		return ports.ProviderConfig{}, ErrInvalid
	}
	stores, err := application.base.bindOperation(ctx, scope, PermissionSend)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	current, err := stores.Providers().LoadProviderConfig(ctx, input.ID)
	if err != nil {
		return ports.ProviderConfig{}, ErrNotFound
	}
	if current.Revision != input.ExpectedRevision ||
		!validateProviderMetadata(current.Kind, input.Metadata) {
		return ports.ProviderConfig{}, ErrConflict
	}
	versionID, err := application.base.nextID(application.base.clock.Now())
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	current.VersionID, current.Name, current.Enabled = versionID, input.Name, input.Enabled
	current.Metadata, current.Revision = input.Metadata, current.Revision+1
	current.MetadataDigest = providerMetadataDigest(current.Kind, input.Metadata)
	if err = stores.Providers().UpdateProviderConfig(ctx, current,
		input.ExpectedRevision); err != nil {
		return ports.ProviderConfig{}, ErrConflict
	}
	if err = stores.Evidence().AppendAudit(ctx, "notifications.provider.update",
		input.ID, scopeActor(scope), current.Revision); err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	return current, nil
}
