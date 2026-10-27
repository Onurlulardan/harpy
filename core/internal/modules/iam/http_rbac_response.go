package iam

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type rbacEnvelope struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type rbacPageData struct {
	Items      any            `json:"items"`
	Pagination rbacPagination `json:"pagination"`
}

type rbacPagination struct {
	HasMore     bool    `json:"has_more"`
	NextAfterID *string `json:"next_after_id,omitempty"`
}

type rbacDeleteData struct {
	DeletedCount int `json:"deleted_count"`
}

func (handler *RBACHTTP) writeRecord(
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
	handler.writeJSON(w, r, status, rbacEnvelope{Success: true, Data: view})
}

func (handler *RBACHTTP) writePage(
	w http.ResponseWriter,
	r *http.Request,
	items any,
	hasMore bool,
	nextAfterID *string,
) {
	handler.writeJSON(w, r, http.StatusOK, rbacEnvelope{
		Success: true,
		Data: rbacPageData{
			Items: items,
			Pagination: rbacPagination{
				HasMore: hasMore, NextAfterID: nextAfterID,
			},
		},
	})
}

func (handler *RBACHTTP) writeDelete(w http.ResponseWriter, r *http.Request) {
	handler.writeJSON(w, r, http.StatusOK, rbacEnvelope{
		Success: true, Data: rbacDeleteData{DeletedCount: 1},
	})
}

func (handler *RBACHTTP) writeJSON(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	payload any,
) {
	if err := httpwire.WriteJSON(
		w, status, authHTTPCorrelationID(r), payload, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM RBAC response: %w", err))
	}
}
