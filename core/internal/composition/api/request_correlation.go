package api

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

func requestCorrelationResolver() func(*http.Request) string {
	return func(request *http.Request) string {
		if request != nil {
			if value, found := requestmeta.CorrelationIDFromContext(request.Context()); found {
				return value.String()
			}
		}
		return requestmeta.EntropyFailureCorrelationID.String()
	}
}
