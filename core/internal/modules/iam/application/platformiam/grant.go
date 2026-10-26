package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) Grant(
	ctx context.Context, scope platformtx.TenantScope, command platformdomain.OperatorGrant,
) (platformdomain.MutationResult, error) {
	if service == nil || command.Validate(service.clock.Now()) != nil ||
		!commandScope(ctx, scope, command.Evidence.ActorID) {
		return platformdomain.MutationResult{}, platformdomain.ErrInvalidCommand
	}
	result, err := service.store.Grant(ctx, scope, command)
	if err != nil {
		return platformdomain.MutationResult{}, err
	}
	if result.Validate(command.TargetUserID) != nil {
		return platformdomain.MutationResult{}, platformdomain.ErrStorage
	}
	return result, nil
}
