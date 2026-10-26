package documentshttp

import "net/http"

func (handler *Handler) deleteFolder(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "folder_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchDocument(w, r, permissionFolderDelete, operationFolderDelete, id, "")
}
