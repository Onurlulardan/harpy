package menuhttp

import "net/http"

type menuOverrideDocument struct {
	ExpectedRevision uint64  `json:"expected_revision"`
	Label            *string `json:"label"`
	Order            *int    `json:"order"`
	Visible          *bool   `json:"visible"`
}

func (handler *Handler) UpdateMenuOverrideRoute() Route {
	return route(http.MethodPatch, "/api/catalog/menu/items/{item_id}/override", handler.updateMenuOverride)
}

func (handler *Handler) updateMenuOverride(writer http.ResponseWriter, request *http.Request) {
	id, err := pathV7(request, "item_id")
	var document menuOverrideDocument
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
	if err != nil || document.ExpectedRevision < 1 ||
		document.Label == nil && document.Order == nil && document.Visible == nil {
		handler.failure(writer, request, ErrInvalid)
		return
	}
	access, err := handler.authorize(request, PermissionMenuManage, "catalog.menu.update", "")
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	value, err := handler.app.UpdateMenuOverride(request.Context(), access, MenuOverrideInput{
		ItemID: id, ExpectedRevision: document.ExpectedRevision, Label: document.Label,
		Order: document.Order, Visible: document.Visible, IdempotencyKey: key,
	})
	if err != nil {
		handler.failure(writer, request, err)
		return
	}
	handler.success(writer, request, http.StatusOK, value)
}
