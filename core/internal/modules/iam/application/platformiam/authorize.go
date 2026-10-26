package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) Authorize(
	ctx context.Context, scope platformtx.TenantScope, query platformdomain.EvaluateQuery,
) (platformdomain.Decision, error) {
	if service == nil || query.Validate() != nil || !queryScope(ctx, scope, query.ActorID) {
		return platformdomain.Decision{}, platformdomain.ErrInvalidQuery
	}
	decision, err := service.store.Authorize(ctx, scope, query)
	if err != nil {
		return platformdomain.Decision{}, err
	}
	if decision.Validate(query) != nil {
		return platformdomain.Decision{}, platformdomain.ErrStorage
	}
	return decision, nil
}
