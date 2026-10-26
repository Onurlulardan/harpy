package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) PurgeSessions(
	ctx context.Context, scope platformtx.TenantScope, command SessionRevokeAllCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.PurgeSessions(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.session.revoke_all", "session", result))
}
