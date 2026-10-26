package menuhttp

import (
	"net/http"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
)

type menuReplaceDocument struct {
	ExpectedRevision uint64                        `json:"expected_revision"`
	Items            []cataloghttp.MenuReplaceItem `json:"items"`
}

func (handler *Handler) ReplaceMenuRoute() Route {
	return route(http.MethodPut, "/api/catalog/menu", handler.replaceMenu)
}

func (handler *Handler) replaceMenu(writer http.ResponseWriter, request *http.Request) {
	var document menuReplaceDocument
	err := requireNoQuery(request)
	if err == nil {
		err = decodeBody(request, &document)
	}
	key := ""
	if err == nil {
		key, err = idempotencyKey(request)
	}
	if err != nil || document.ExpectedRevision < 1 || document.Items == nil {
		handler.failure(writer, request, ErrInvalid)
		return
	}
	access, err := handler.authorize(request, PermissionMenuManage, "catalog.menu.update", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.ReplaceMenu(request.Context(), access, cataloghttp.MenuReplaceInput{
		ExpectedRevision: document.ExpectedRevision, Items: document.Items, IdempotencyKey: key,
	})
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
