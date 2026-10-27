package lifecyclehttp

import "net/http"

type upgradeDocument struct {
	Version string `json:"version"`
}

func (handler *Handler) UpgradeRoute() Route {
	return route(http.MethodPost, "/api/catalog/installations/{installation_id}/upgrades", handler.upgrade)
}

func (handler *Handler) upgrade(writer http.ResponseWriter, request *http.Request) {
	id, err := pathV7(request, "installation_id")
	var document upgradeDocument
	if err == nil {
		err = requireNoQuery(request)
	}
	if err == nil {
		err = decodeBody(request, &document)
	}
	key := ""
	if err == nil {
		key, err = idempotencyKey(request)
	}
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	access, err := handler.authorize(request, PermissionUpgrade, "catalog.installation.upgrade", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.Upgrade(request.Context(), access, id, document.Version, key)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusAccepted, value)
}
