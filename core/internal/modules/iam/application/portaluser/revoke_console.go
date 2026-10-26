package portaluser

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) RevokeConsole(
	ctx context.Context, scope platformtx.TenantScope, command RevisionCommand,
) (MutationResult, error) {
	if service == nil || !scopeUsable(ctx, scope, command.Evidence, platformtx.Command) ||
		command.UserID == [16]byte{} || command.ExpectedRevision < 1 {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.RevokeConsole(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	if !validResult(result, command.UserID) {
		return MutationResult{}, ErrUnavailable
	}
	return result, nil
}
