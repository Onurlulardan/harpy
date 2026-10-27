package policy

import (
	"context"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) EvaluateQuery(
	ctx context.Context,
	scope transaction.TenantScope,
	request restauthorization.PolicyRequest,
) (restauthorization.PolicyDecision, error) {
	return service.Authorize(ctx, scope, request)
}
