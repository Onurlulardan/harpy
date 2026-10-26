package documentshttp

import "net/http"

func (handler *Handler) getUpload(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "upload_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionItemRead, operationUploadGet, id, "")
}
