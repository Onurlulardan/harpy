package queryexecutionhttp

import (
	"errors"
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryexecutionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/queryexecution"
)

func (handler *Handler) Execute(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionQueryExecute, httptransport.OperationQueryExecute)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	if request.URL.RawQuery != "" {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	document, err := httptransport.Decode[executeDocument](writer, request)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	batch, ok := parse(document)
	if !ok {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	defer clearBatch(batch)
	result, err := handler.application.ExecuteQueryBatch(request.Context(), access, batch)
	if err != nil {
		if errors.Is(err, queryexecutionapp.ErrInvalid) {
			err = httptransport.ErrInvalid
		} else {
			err = httptransport.ErrUnavailable
		}
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, present(result))
}

func clearBatch(batch queryexecutionapp.Request) {
	for index := range batch.Items {
		clear(batch.Items[index].Parameters)
	}
}

func parse(document executeDocument) (queryexecutionapp.Request, bool) {
	items := make([]queryexecutionapp.Item, len(document.Queries))
	for index, query := range document.Queries {
		queryID, queryOK := httptransport.WireV7(query.QueryID)
		versionID, versionOK := httptransport.WireV7(query.VersionID)
		if !queryOK || !versionOK {
			return queryexecutionapp.Request{}, false
		}
		items[index] = queryexecutionapp.Item{
			DataSourceID: query.DataSourceID,
			QueryID:      queryID,
			VersionID:    versionID,
			Parameters:   append([]byte(nil), query.Params...),
		}
	}
	return queryexecutionapp.Request{Items: items}, true
}
