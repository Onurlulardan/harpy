package reports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) GetRevenue(
	ctx context.Context,
	scope transaction.GlobalScope,
	request api.QueryRequest,
) (RevenueResult, error) {
	if service == nil || service.engine == nil {
		return RevenueResult{}, ErrContract
	}
	if err := validateRevenueShape(request); err != nil {
		return RevenueResult{}, err
	}
	result, err := service.engine.ExecuteAggregate(ctx, scope, RevenueResource, request)
	if err != nil {
		return RevenueResult{}, err
	}
	return newRevenueResult(result.Groups(), result.Limit()), nil
}
