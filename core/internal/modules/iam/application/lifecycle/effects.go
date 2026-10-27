package lifecycle

import (
	"context"
	"errors"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) finishTenant(
	ctx context.Context, scope platformtx.TenantScope, result MutationResult,
	change Change,
) (MutationResult, error) {
	if result.Validate() != nil {
		return MutationResult{}, ErrUnavailable
	}
	if result.Replayed {
		return result, nil
	}
	if change.Validate() != nil || service.effects.WriteTenantChange(ctx, scope, change) != nil {
		return MutationResult{}, errors.Join(ErrEffects, ErrUnavailable)
	}
	return result, nil
}

func (service *Service) finishCredential(
	ctx context.Context, scope platformtx.CredentialTenantScope,
	result MutationResult, change Change,
) (MutationResult, error) {
	if result.Validate() != nil {
		return MutationResult{}, ErrUnavailable
	}
	if result.Replayed {
		return result, nil
	}
	if change.Validate() != nil ||
		service.effects.WriteCredentialChange(ctx, scope, change) != nil {
		return MutationResult{}, errors.Join(ErrEffects, ErrUnavailable)
	}
	return result, nil
}
