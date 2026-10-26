package publicationhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Get(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionPublicationRead, httptransport.OperationPublicationGet)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	releaseID, idOK := httptransport.PathV7(request, "release_id")
	if !idOK || request.URL.RawQuery != "" || !httptransport.BodyAbsent(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	result, err := handler.application.GetPublication(request.Context(), access, releaseID)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, present(result))
}
