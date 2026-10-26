package lifecyclehttp

import "net/http"

func (handler *Handler) ReadOperationRoute() Route {
	return route(http.MethodGet, "/api/catalog/operations/{operation_id}", handler.readOperation)
}

func (handler *Handler) readOperation(writer http.ResponseWriter, request *http.Request) {
	id, err := pathV7(request, "operation_id")
	if err == nil {
		err = requireNoQuery(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionRead, "catalog.operation.get", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.GetOperation(request.Context(), access, id)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
