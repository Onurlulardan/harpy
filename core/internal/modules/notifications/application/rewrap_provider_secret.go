package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type RewrapProviderSecretInput struct {
	ID                  domain.ID
	TargetKEKRevision   int64
	ExpectedKEKRevision int64
}

func (application *ProviderApplication) RewrapProviderSecret(ctx context.Context,
	scope api.Scope, input RewrapProviderSecretInput,
) error {
	if application == nil || !input.ID.Valid() || input.TargetKEKRevision < 1 ||
		input.ExpectedKEKRevision < 1 ||
		input.TargetKEKRevision == input.ExpectedKEKRevision {
		return ErrInvalid
	}
	stores, err := application.base.bindOperation(ctx, scope, PermissionSend)
	if err != nil {
		return err
	}
	config, err := stores.Providers().LoadProviderConfig(ctx, input.ID)
	if err != nil {
		return ErrNotFound
	}
	secret, err := stores.Providers().LoadProviderSecret(ctx, config.SecretID)
	if err != nil {
		return ErrNotFound
	}
	if secret.KEKRevision != input.ExpectedKEKRevision {
		return ErrConflict
	}
	rewrapped, err := application.sealer.Rewrap(ctx, secret, input.TargetKEKRevision)
	if err != nil {
		return ErrUnavailable
	}
	if err = stores.Providers().RewrapProviderSecret(ctx, rewrapped,
		input.ExpectedKEKRevision); err != nil {
		return ErrConflict
	}
	return stores.Evidence().AppendAudit(ctx, "notifications.provider.secret.rewrap",
		input.ID, scopeActor(scope), config.Revision)
}
