package documentshttp

import "net/http"

func (handler *Handler) putUploadContent(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "upload_id")
	if !valid || !contentRequestValid(r) {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	access, err := handler.accessFor(r, permissionItemVersion, operationUploadPut)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	response, err := handler.application.Dispatch(r.Context(), Request{Access: access,
		Operation: operationUploadPut, ResourceID: id, Content: r.Body})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeResponse(w, r, response)
}
