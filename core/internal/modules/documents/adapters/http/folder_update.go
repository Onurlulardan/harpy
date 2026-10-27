package documentshttp

import "net/http"

func (handler *Handler) updateFolder(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "folder_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchDocument(w, r, permissionFolderUpdate, operationFolderUpdate, id, "")
}
