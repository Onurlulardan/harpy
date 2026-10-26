package documentshttp

import "net/http"

func (handler *Handler) restoreItem(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "document_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchDocument(w, r, permissionItemDelete, operationItemRestore, id, "")
}
