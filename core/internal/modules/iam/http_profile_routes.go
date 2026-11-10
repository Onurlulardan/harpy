package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *ProfileHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/profile", Handler: http.HandlerFunc(handler.getSelf)},
		{Method: http.MethodPut, Pattern: "/api/profile", Handler: http.HandlerFunc(handler.updateSelf)},
	}
}
