package queryexecutionhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

const queryExecutionsPath = "/api/analytics/query-executions"

func (handler *Handler) Routes() []httptransport.Route {
	return []httptransport.Route{{
		Method:  http.MethodPost,
		Pattern: queryExecutionsPath,
		Handler: http.HandlerFunc(handler.Execute),
	}}
}
