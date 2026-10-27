package api

import (
	"net/http"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) writeRecordHTTPQuery(
	w http.ResponseWriter,
	r *http.Request,
	view systemResourceRecordHTTPQueryView,
) {
	if view.Items == nil {
		handler.writeRecordHTTPFailure(w, r, systemresources.ErrUnavailable)
		return
	}
	handler.writeRecordHTTPJSON(w, r, http.StatusOK, systemResourceRecordHTTPSuccess{
		Success: true,
		Data: systemResourceRecordHTTPQueryData{
			Items: view.Items,
			Pagination: systemResourceRecordHTTPQueryPagination{
				HasMore: view.HasMore, NextCursor: view.NextCursor,
				ExactCount: view.ExactCount,
			},
		},
	})
}
