package documentshttp

import "net/http"

func (handler *Handler) listItems(w http.ResponseWriter, r *http.Request) {
	handler.dispatchRead(w, r, permissionItemRead, operationItemList, "", "")
}

func (handler *Handler) batchItems(w http.ResponseWriter, r *http.Request) {
	handler.dispatchDocument(w, r, permissionItemRead, operationItemBatch, "", "")
}

func (handler *Handler) getItem(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "document_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionItemRead, operationItemGet, id, "")
}

func (handler *Handler) listVersions(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "document_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionItemRead, operationVersionList, id, "")
}

func (handler *Handler) latestVersion(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "document_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchRead(w, r, permissionItemRead, operationVersionLatest, id, "")
}
