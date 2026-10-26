package platformops

import (
	"net/http"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func (handler *HTTPHandler) listAlertEvents(w http.ResponseWriter, r *http.Request) {
	environment, _, err := handler.alertEnvironment(r, "platform.operations.alerts.read")
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	filter, err := alertEventFilter(r)
	if err == nil {
		page, readErr := handler.alerts.ListAlertEvents(
			r.Context(), environment.EnvironmentID, filter)
		err = readErr
		if err == nil {
			views := make([]alertEventView, len(page.Items))
			for index, event := range page.Items {
				views[index] = alertEventAPI(event)
			}
			handler.writePage(w, r, views, platformOpsPageMeta{Total: page.Total,
				Page: page.Page, Limit: page.Limit, TotalPages: page.TotalPages})
			return
		}
	}
	handler.writeFailure(w, r, err)
}

func alertEventFilter(r *http.Request) (platformopsapi.AlertEventFilter, error) {
	if r == nil {
		return platformopsapi.AlertEventFilter{}, errHTTPRequestInvalid
	}
	query := r.URL.Query()
	allowed := map[string]bool{"page": true, "limit": true, "service_name": true, "status": true}
	for key, values := range query {
		if !allowed[key] || len(values) != 1 {
			return platformopsapi.AlertEventFilter{}, errHTTPRequestInvalid
		}
	}
	page, err := queryInt(query.Get("page"))
	if err != nil {
		return platformopsapi.AlertEventFilter{}, err
	}
	limit, err := queryInt(query.Get("limit"))
	if err != nil {
		return platformopsapi.AlertEventFilter{}, err
	}
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 50
	}
	return platformopsapi.AlertEventFilter{Page: page, Limit: limit,
		ServiceName: query.Get("service_name"), Status: query.Get("status")}, nil
}
