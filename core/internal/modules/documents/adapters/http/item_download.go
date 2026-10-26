package documentshttp

import "net/http"

func (handler *Handler) downloadItem(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "document_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionItemDownload, operationDownload, id, "")
}

func (handler *Handler) downloadVersion(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "document_id")
	version, versionValid := pathID(r, "version_id")
	if !valid || !versionValid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionItemDownload,
		operationVersionContent, id, version)
}
