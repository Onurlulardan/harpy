package queryexecutionhttp

import (
	"context"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryexecutionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/queryexecution"
)

type Application interface {
	ExecuteQueryBatch(
		context.Context,
		httptransport.Access,
		queryexecutionapp.Request,
	) (queryexecutionapp.Result, error)
}
