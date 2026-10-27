package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type systemResourceRecordHTTPSuccess struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type systemResourceRecordHTTPDeleteEnvelope struct {
	Success      bool `json:"success"`
	DeletedCount int  `json:"deleted_count"`
}

type systemResourceRecordHTTPQueryData struct {
	Items      []json.RawMessage                       `json:"items"`
	Pagination systemResourceRecordHTTPQueryPagination `json:"pagination"`
}

type systemResourceRecordHTTPQueryPagination struct {
	Limit      int    `json:"limit,omitempty"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
	ExactCount *int64 `json:"exact_count,omitempty"`
}

func (handler *systemResourceRecordHTTPHandler) writeRecordHTTPMutation(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	view systemResourceRecordHTTPView,
	resource systemresources.Resource,
) {
	if !validSystemResourceRecordHTTPView(view) {
		handler.writeRecordHTTPFailure(w, r, systemresources.ErrUnavailable)
		return
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, view.Revision))
	if status == http.StatusCreated {
		w.Header().Set("Location", fmt.Sprintf("/api/%s/%s", string(resource), view.ID))
	}
	handler.writeRecordHTTPJSON(w, r, status, systemResourceRecordHTTPSuccess{
		Success: true, Data: view.Data,
	})
}

func (handler *systemResourceRecordHTTPHandler) writeRecordHTTPRead(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	view systemResourceRecordHTTPView,
	location string,
) {
	if !validSystemResourceRecordHTTPView(view) {
		handler.writeRecordHTTPFailure(w, r, systemresources.ErrUnavailable)
		return
	}
	if location != "" {
		w.Header().Set("Location", location)
	}
	w.Header().Set("ETag", fmt.Sprintf(`"revision-%d"`, view.Revision))
	handler.writeRecordHTTPJSON(w, r, status, systemResourceRecordHTTPSuccess{
		Success: true, Data: view.Data,
	})
}

func validSystemResourceRecordHTTPView(view systemResourceRecordHTTPView) bool {
	if view.ID == "" || view.Revision < 1 || len(view.Data) == 0 {
		return false
	}
	var id string
	var revision int64
	return json.Unmarshal(view.Data["id"], &id) == nil && id == view.ID &&
		json.Unmarshal(view.Data["revision"], &revision) == nil &&
		revision == view.Revision
}

func (handler *systemResourceRecordHTTPHandler) writeRecordHTTPDelete(
	w http.ResponseWriter, r *http.Request,
) {
	handler.writeRecordHTTPJSON(w, r, http.StatusOK, systemResourceRecordHTTPDeleteEnvelope{
		Success: true, DeletedCount: 1,
	})
}

func (handler *systemResourceRecordHTTPHandler) writeRecordHTTPJSON(
	w http.ResponseWriter, r *http.Request, status int, payload any,
) {
	if err := httpwire.WriteJSON(
		w, status, layoutHTTPCorrelationID(r), payload, false,
	); err != nil {
		handler.report(errors.New("system resource record HTTP response write failed"))
	}
}
