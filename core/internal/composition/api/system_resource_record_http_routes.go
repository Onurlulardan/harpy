package api

import (
	"net/http"

	publicresource "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *systemResourceRecordHTTPHandler) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	resources := []publicresource.Resource{
		publicresource.Users,
		publicresource.Roles,
		publicresource.Organizations,
	}
	result := make([]httpedge.Route, 0, len(resources)*5)
	for _, resource := range resources {
		collection := "/api/" + string(resource)
		item := collection + "/{id}"
		result = append(result,
			httpedge.Route{Method: http.MethodGet, Pattern: collection,
				Handler: handler.list(resource)},
			httpedge.Route{Method: http.MethodPost, Pattern: collection,
				Handler: handler.create(resource)},
			httpedge.Route{Method: http.MethodGet, Pattern: item,
				Handler: handler.get(resource)},
			httpedge.Route{Method: http.MethodPut, Pattern: item,
				Handler: handler.update(resource)},
			httpedge.Route{Method: http.MethodDelete, Pattern: item,
				Handler: handler.delete(resource)},
		)
	}
	return result
}
