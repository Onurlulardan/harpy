package reports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) ExecuteLegacyEvents(
	ctx context.Context,
	scope transaction.GlobalScope,
	request api.QueryRequest,
) (RecordResult, error) {
	if service == nil || service.engine == nil || request.Mode() != api.QueryModeRecord {
		return RecordResult{}, ErrContract
	}
	result, err := service.engine.Execute(ctx, scope, LegacyEventResource, request)
	if err != nil {
		return RecordResult{}, err
	}
	return newRecordResult(result), nil
}
