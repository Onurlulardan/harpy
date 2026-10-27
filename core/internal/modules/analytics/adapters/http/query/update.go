package queryhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func (handler *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionQueryUpdate, httptransport.OperationQueryUpdate)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	queryID, idOK := httptransport.PathV7(request, "query_id")
	key, keyOK := httptransport.IdempotencyKey(request)
	if !idOK || !keyOK || request.URL.RawQuery != "" {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.Decode[updateDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	defer clear(document.Query)
	defer clear(document.Parameters)
	defer clear(document.Output)
	result, err := handler.application.UpdateQuery(request.Context(), access, queryapp.UpdateRequest{
		QueryID: queryID, ExpectedRootRevision: document.ExpectedRootRevision,
		ExpectedVersionRevision: document.ExpectedVersionRevision, Name: document.Name,
		QueryDocument: document.Query, ParameterDocument: document.Parameters,
		OutputDocument: document.Output, IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, present(result))
}
