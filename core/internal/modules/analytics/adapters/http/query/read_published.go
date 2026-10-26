package queryhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) ReadPublished(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(
		request,
		httptransport.PermissionQueryRead,
		httptransport.OperationQueryGetPublished,
	)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	queryID, idOK := httptransport.PathV7(request, "query_id")
	if !idOK || request.URL.RawQuery != "" || !httptransport.BodyAbsent(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.GetPublishedQuery(
		request.Context(),
		access,
		queryID,
	)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, presentPublished(result))
}
