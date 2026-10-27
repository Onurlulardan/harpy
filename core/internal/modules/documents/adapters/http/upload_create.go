package documentshttp

import "net/http"

func (handler *Handler) createUpload(w http.ResponseWriter, r *http.Request) {
	handler.dispatchDocument(w, r, permissionItemCreate, operationUploadCreate, "", "")
}
