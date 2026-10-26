package queryhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionQueryRead, httptransport.OperationQueryGet)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	queryID, idOK := httptransport.PathV7(request, "query_id")
	if !idOK || request.URL.RawQuery != "" || !httptransport.BodyAbsent(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.GetQuery(request.Context(), access, queryID)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, present(result))
}
