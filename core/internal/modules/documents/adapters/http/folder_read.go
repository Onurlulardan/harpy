package documentshttp

import "net/http"

func (handler *Handler) listFolders(w http.ResponseWriter, r *http.Request) {
	handler.dispatchRead(w, r, permissionFolderRead, operationFolderList, "", "")
}

func (handler *Handler) getFolder(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "folder_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionFolderRead, operationFolderGet, id, "")
}

func (handler *Handler) listFolderContents(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "folder_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionFolderRead, operationFolderContents, id, "")
}
