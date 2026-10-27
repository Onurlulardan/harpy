package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) SuspendAuthenticator(
	ctx context.Context, scope platformtx.TenantScope, command RevisionCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.ID == uuid.Nil || command.ExpectedRevision < 1 {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.SuspendAuthenticator(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.authenticator.suspend", "authenticator", result))
}
