package menuhttp

import (
	"encoding/json"
	"net/http"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
)

type createMenuConfigDocument struct {
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

func (handler *Handler) CreateMenuConfigRoute() Route {
	return route(http.MethodPost, "/api/form-engine/menu/configs", handler.createMenuConfig)
}

func (handler *Handler) createMenuConfig(writer http.ResponseWriter, request *http.Request) {
	var document createMenuConfigDocument
	err := requireNoQuery(request)
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
	value, err := handler.app.CreateMenuConfig(request.Context(), access,
		cataloghttp.MenuConfigWriteInput{Name: document.Name, Config: document.Config,
			Active: true, IdempotencyKey: key})
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusCreated, value)
}
