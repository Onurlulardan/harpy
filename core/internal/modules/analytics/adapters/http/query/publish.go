package queryhttp

import "net/http"

func (handler *Handler) Publish(writer http.ResponseWriter, request *http.Request) {
	access, input, err := handler.publicationInput(
		writer, request, operationQueryPublish,
	)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	result, err := handler.application.PublishQuery(request.Context(), access, input)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusCreated, present(result))
}
