package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) ChangePassword(
	ctx context.Context, scope platformtx.TenantScope, command PasswordChangeCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.UserID == uuid.Nil || command.ExpectedCredentialRevision < 1 ||
		command.Salt == [16]byte{} || command.Tag == [32]byte{} {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.ChangePassword(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result, changeFrom(command.Evidence,
		"iam.password.change", "password_credential", result))
}
