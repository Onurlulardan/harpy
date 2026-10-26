package queryhttp

import "net/http"

func (handler *Handler) Revalidate(writer http.ResponseWriter, request *http.Request) {
	access, input, err := handler.publicationInput(
		writer, request, operationQueryRevalidate,
	)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	result, err := handler.application.RevalidateQuery(request.Context(), access, input)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	handler.runtime.JSON(writer, request, http.StatusCreated, present(result))
}
