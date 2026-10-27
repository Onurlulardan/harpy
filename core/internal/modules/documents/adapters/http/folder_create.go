package documentshttp

import "net/http"

func (handler *Handler) createFolder(w http.ResponseWriter, r *http.Request) {
	handler.dispatchDocument(w, r, permissionFolderCreate, operationFolderCreate, "", "")
}
