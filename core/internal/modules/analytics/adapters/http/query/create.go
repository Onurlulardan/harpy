package queryhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
)

func (handler *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionQueryCreate, httptransport.OperationQueryCreate)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	key, keyOK := httptransport.IdempotencyKey(request)
	if request.URL.RawQuery != "" || !keyOK {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.Decode[createDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	defer clear(document.Query)
	defer clear(document.Parameters)
	defer clear(document.Output)
	result, err := handler.application.CreateQuery(request.Context(), access, queryapp.CreateRequest{
		Name: document.Name, QueryDocument: document.Query,
		ParameterDocument: document.Parameters, OutputDocument: document.Output,
		IdempotencyKey: key})
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	writer.Header().Set("Location", queriesPath+"/"+result.QueryID.String())
	handler.runtime.JSON(writer, request, http.StatusCreated, present(result))
}
