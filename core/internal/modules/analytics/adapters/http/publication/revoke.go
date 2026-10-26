package publicationhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Revoke(writer http.ResponseWriter, request *http.Request) {
	access, input, err := handler.transition(writer, request, httptransport.OperationPublicationRevoke)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	result, err := handler.application.RevokePublication(request.Context(), access, input)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusOK, present(result))
}
