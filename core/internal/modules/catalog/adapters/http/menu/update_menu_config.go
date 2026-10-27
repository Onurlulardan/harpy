package menuhttp

import (
	"encoding/json"
	"net/http"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
)

type updateMenuConfigDocument struct {
	Name     string          `json:"name"`
	Config   json.RawMessage `json:"config"`
	IsActive bool            `json:"is_active"`
}

func (handler *Handler) UpdateMenuConfigRoute() Route {
	return route(http.MethodPut, "/api/form-engine/menu/configs/{id}", handler.updateMenuConfig)
}

func (handler *Handler) updateMenuConfig(writer http.ResponseWriter, request *http.Request) {
	id, err := pathV7(request, "id")
	var document updateMenuConfigDocument
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
	if err != nil || document.Name == "" || len(document.Config) == 0 {
		handler.failure(writer, request, ErrInvalid)
		return
	}
	access, err := handler.authorize(request, PermissionMenuManage, "catalog.menu.update", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.UpdateMenuConfig(request.Context(), access, id,
		cataloghttp.MenuConfigWriteInput{Name: document.Name, Config: document.Config,
			Active: document.IsActive, IdempotencyKey: key})
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
