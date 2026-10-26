package analytics

import (
	"context"

	analyticshttp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/http"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	queryexecutionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/queryexecution"
)

func (runtime *HTTPRuntime) ExecuteQueryBatch(
	ctx context.Context,
	access analyticshttp.Access,
	request queryexecutionapp.Request,
) (queryexecutionapp.Result, error) {
	return queryexecutionapp.Execute(ctx, request, queryBatchRunner{
		runtime: runtime,
		access:  access,
	})
}

type queryBatchRunner struct {
	runtime *HTTPRuntime
	access  analyticshttp.Access
}

func (runner queryBatchRunner) Run(
	ctx context.Context,
	item queryexecutionapp.Item,
) ([]byte, error) {
	result, err := runner.runtime.ExecuteQuery(ctx, runner.access, queryapp.ExecuteRequest{
		QueryID: item.QueryID, VersionID: item.VersionID, Parameters: item.Parameters,
	})
	return result.Body, err
}
