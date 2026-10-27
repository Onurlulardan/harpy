package documentshttp

import "net/http"

func (handler *Handler) cancelUpload(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "upload_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchDocument(w, r, permissionItemVersion, operationUploadCancel, id, "")
}
