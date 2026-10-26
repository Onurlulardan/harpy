package queryhttp

import (
	"encoding/json"
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func (handler *Handler) Execute(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionQueryExecute, httptransport.OperationQueryExecute)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	queryID, idOK := httptransport.PathV7(request, "query_id")
	if !idOK || request.URL.RawQuery != "" {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.Decode[executeDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	defer clear(document.Params)
	versionID, versionOK := httptransport.WireV7(document.VersionID)
	if !versionOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.ExecuteQuery(request.Context(), access, queryapp.ExecuteRequest{
		QueryID: queryID, VersionID: versionID, Parameters: document.Params})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, json.RawMessage(result.Body))
}
