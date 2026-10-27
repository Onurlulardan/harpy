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
		{Method: http.MethodGet, Pattern: "/api/iam/profile-schemas/{resource}/active", Handler: http.HandlerFunc(handler.getActiveSchema)},
		{Method: http.MethodGet, Pattern: "/api/iam/profile-schemas/{resource}/{schema_id}", Handler: http.HandlerFunc(handler.getSchema)},
		{Method: http.MethodPost, Pattern: "/api/iam/profile-schemas/{resource}/publish", Handler: http.HandlerFunc(handler.publishSchema)},
		{Method: http.MethodPost, Pattern: "/api/iam/profile-schemas/{resource}/backfills/{plan_id}/recover", Handler: http.HandlerFunc(handler.recoverBackfill)},
		{Method: http.MethodPost, Pattern: "/api/iam/profile-schemas/{resource}/fields/{field_id}/retire", Handler: http.HandlerFunc(handler.retireField)},
	}
}
