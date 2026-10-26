package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

type systemTransferV1Definition struct {
	method, suffix, successor string
}

func (handler *SystemTransferHTTP) systemTransferV1Routes() []httpedge.Route {
	definitions := []systemTransferV1Definition{
		{http.MethodGet, "/export", "/record-exports"},
		{http.MethodPost, "/import/analyze", "/record-import-sources"},
		{http.MethodPost, "/import/execute", "/record-imports"},
		{http.MethodGet, "/import/jobs", "/record-imports"},
		{http.MethodGet, "/import/jobs/{job_id}", "/record-imports/{import_id}"},
		{http.MethodGet, "/import/jobs/{job_id}/rows", "/record-imports/{import_id}/report"},
		{http.MethodPost, "/import/jobs/{job_id}/resume", "/record-imports/{import_id}"},
		{http.MethodPost, "/import/jobs/{job_id}/cancel", "/record-imports/{import_id}"},
		{http.MethodGet, "/import/jobs/{job_id}/report", "/record-imports/{import_id}/report"},
		{http.MethodPost, "/import/sessions/{session_id}/cancel", "/record-imports/{import_id}"},
	}
	resources := []string{"users", "roles", "organizations"}
	routes := make([]httpedge.Route, 0, len(resources)*len(definitions))
	for _, resource := range resources {
		for _, definition := range definitions {
			pattern := "/api/" + resource + definition.suffix
			successor := "/api/iam/system-resources/" + resource + definition.successor
			routes = append(routes, systemTransferV1Route(
				handler, definition.method, pattern, successor))
		}
	}
	return routes
}

func systemTransferV1Route(handler *SystemTransferHTTP, method, pattern,
	successor string) httpedge.Route {
	return httpedge.Route{Method: method, Pattern: pattern,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.systemTransferUpgradeRequired(w, r, successor)
		})}
}
