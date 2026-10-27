package lifecyclehttp

import "net/http"

func (handler *Handler) RestoreRoute() Route {
	return route(http.MethodPost, "/api/catalog/installations/{installation_id}/restorations", handler.restore)
}

func (handler *Handler) restore(writer http.ResponseWriter, request *http.Request) {
	id, err := pathV7(request, "installation_id")
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
	access, err := handler.authorize(request, PermissionRestore, "catalog.installation.restore", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.Restore(request.Context(), access, id, key)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusAccepted, value)
}
