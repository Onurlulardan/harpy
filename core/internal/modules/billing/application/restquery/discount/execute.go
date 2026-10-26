package discount

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) Execute(
	ctx context.Context, scope transaction.GlobalScope, request Request,
) (Result, error) {
	value, err := request.Value()
	if service == nil || service.engine == nil || err != nil {
		return Result{}, ErrContract
	}
	result, err := service.engine.Execute(ctx, scope, ResourceName, value)
	if err != nil {
		return Result{}, err
	}
	transport := ""
	if result.HasMore() {
		transport = request.Transport()
	}
	return newResult(result, transport), nil
}
