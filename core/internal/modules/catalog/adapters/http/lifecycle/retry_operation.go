package lifecyclehttp

import "net/http"

func (handler *Handler) RetryOperationRoute() Route {
	return route(http.MethodPost, "/api/catalog/operations/{operation_id}/retries", handler.retryOperation)
}

func (handler *Handler) retryOperation(writer http.ResponseWriter, request *http.Request) {
	id, err := pathV7(request, "operation_id")
	if err == nil {
		err = requireNoQuery(request)
	}
	if err == nil {
		err = requireNoBody(request)
	}
	key := ""
	if err == nil {
		key, err = idempotencyKey(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionInstall, "catalog.operation.retry", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.RetryOperation(request.Context(), access, id, key)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusAccepted, value)
}
