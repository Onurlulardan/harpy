package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) AddIdentifier(
	ctx context.Context, scope platformtx.TenantScope, command IdentifierAddCommand,
) (MutationResult, error) {
	if !service.Valid() || !tenantCommandUsable(ctx, scope, command.Evidence) ||
		command.ID == uuid.Nil || command.UserID == uuid.Nil || command.Kind == "" ||
		command.DisplayValue == "" || command.NormalizedValue == "" ||
		command.NormalizedSHA256 == [32]byte{} {
		return MutationResult{}, ErrInvalid
	}
	result, err := service.store.AddIdentifier(ctx, scope, command)
	if err != nil {
		return MutationResult{}, err
	}
	return service.finishTenant(ctx, scope, result,
		changeFrom(command.Evidence, "iam.identifier.add", "identifier", result))
}
