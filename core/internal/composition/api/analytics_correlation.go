package api

import (
	"net/http"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func analyticsCorrelationResolver() analyticsmodule.HTTPCorrelationResolver {
	return func(request *http.Request) string {
		if request != nil {
			if value, found := requestmeta.CorrelationIDFromContext(request.Context()); found {
				return value.String()
			}
		}
		return requestmeta.EntropyFailureCorrelationID.String()
	}
}
