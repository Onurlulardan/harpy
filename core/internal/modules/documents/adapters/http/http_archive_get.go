package documentshttp

import "net/http"

func (handler *Handler) getArchive(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "archive_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionItemDownload, operationArchiveGet, id, "")
}
