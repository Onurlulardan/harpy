package documentshttp

import "net/http"

func (handler *Handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "document_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchDocument(w, r, permissionItemDelete, operationItemDelete, id, "")
}
