package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type RotateProviderSecretInput struct {
	ID               domain.ID
	Secret           []byte
	KEKRevision      int64
	ExpectedRevision int64
}

func (application *ProviderApplication) RotateProviderSecret(ctx context.Context,
	scope api.Scope, input RotateProviderSecretInput,
) (ports.ProviderConfig, error) {
	defer clear(input.Secret)
	if application == nil || !input.ID.Valid() || len(input.Secret) == 0 ||
		len(input.Secret) > 16*1024 || input.KEKRevision < 1 || input.ExpectedRevision < 1 {
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
	if current.Revision != input.ExpectedRevision {
		return ports.ProviderConfig{}, ErrConflict
	}
	now := application.base.clock.Now()
	secretID, err := application.base.nextID(now)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	versionID, err := application.base.nextID(now)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	sealed, err := application.sealer.Seal(ctx, secretID, input.Secret, input.KEKRevision)
	if err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	current.SecretID, current.VersionID = secretID, versionID
	current.Revision++
	if err = stores.Providers().RotateProviderSecret(ctx, current, sealed,
		input.ExpectedRevision); err != nil {
		return ports.ProviderConfig{}, ErrConflict
	}
	if err = stores.Evidence().AppendAudit(ctx, "notifications.provider.secret.rotate",
		input.ID, scopeActor(scope), current.Revision); err != nil {
		return ports.ProviderConfig{}, ErrUnavailable
	}
	return current, nil
}
