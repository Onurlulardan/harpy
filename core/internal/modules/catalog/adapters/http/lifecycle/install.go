package lifecyclehttp

import (
	"net/http"
)

type installDocument struct {
	PackageID string `json:"package_id"`
	Version   string `json:"version"`
}

func (handler *Handler) InstallRoute() Route {
	return route(http.MethodPost, "/api/catalog/installations", handler.install)
}

func (handler *Handler) install(writer http.ResponseWriter, request *http.Request) {
	var document installDocument
	err := requireNoQuery(request)
	if err == nil {
		err = decodeBody(request, &document)
	}
	key := ""
	if err == nil {
		key, err = idempotencyKey(request)
	}
	packageID, packageErr := parsePackageID(document.PackageID)
	if err != nil || packageErr != nil {
		handler.failure(writer, request, ErrInvalid)
		return
	}
	access, err := handler.authorize(request, PermissionInstall, "catalog.installation.install", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.Install(request.Context(), access, packageID, document.Version, key)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusAccepted, value)
}
