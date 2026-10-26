package iam

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type permissionRequestRecordEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type permissionRequestPageData struct {
	Items      []PermissionRequestGroupView `json:"items"`
	Pagination permissionRequestPagination  `json:"pagination"`
}

type permissionRequestPagination struct {
	HasMore     bool    `json:"has_more"`
	NextAfterID *string `json:"next_after_id,omitempty"`
}

type permissionRequestCatalogPageData struct {
	Items      any                         `json:"items"`
	Pagination permissionRequestPagination `json:"pagination"`
}

func (handler *PermissionRequestHTTP) writeRecord(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	revision int64,
	view any,
) {
	if revision < 1 {
		handler.writeFailure(w, r, identityFailureUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, revision))
	handler.writeJSON(w, r, status, permissionRequestRecordEnvelope{
		Success: true, Data: view,
	})
}

func (handler *PermissionRequestHTTP) writePage(
	w http.ResponseWriter,
	r *http.Request,
	view PermissionRequestPageView,
) {
	data := permissionRequestPageData{
		Items: view.Items,
		Pagination: permissionRequestPagination{
			HasMore: view.HasMore, NextAfterID: view.NextAfterID,
		},
	}
	handler.writeJSON(w, r, http.StatusOK, permissionRequestRecordEnvelope{
		Success: true, Data: data,
	})
}

func (handler *PermissionRequestHTTP) writeCatalogResourcePage(
	w http.ResponseWriter,
	r *http.Request,
	view PermissionRequestCatalogResourcePageView,
) {
	handler.writeCatalogPage(w, r, view.Items, view.HasMore, view.NextAfterID)
}

func (handler *PermissionRequestHTTP) writeCatalogActionPage(
	w http.ResponseWriter,
	r *http.Request,
	view PermissionRequestCatalogActionPageView,
) {
	handler.writeCatalogPage(w, r, view.Items, view.HasMore, view.NextAfterID)
}

func (handler *PermissionRequestHTTP) writeCatalogPage(
	w http.ResponseWriter,
	r *http.Request,
	items any,
	hasMore bool,
	nextAfterID *string,
) {
	handler.writeJSON(w, r, http.StatusOK, permissionRequestRecordEnvelope{
		Success: true,
		Data: permissionRequestCatalogPageData{
			Items: items,
			Pagination: permissionRequestPagination{
				HasMore: hasMore, NextAfterID: nextAfterID,
			},
		},
	})
}

func (handler *PermissionRequestHTTP) writeJSON(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	payload any,
) {
	if err := httpwire.WriteJSON(
		w, status, authHTTPCorrelationID(r), payload, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM permission request response: %w", err))
	}
}
