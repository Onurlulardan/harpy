package documentshttp

import "net/http"

func (handler *Handler) createArchive(w http.ResponseWriter, r *http.Request) {
	handler.dispatchDocument(w, r, permissionItemDownload, operationArchiveCreate, "", "")
}
