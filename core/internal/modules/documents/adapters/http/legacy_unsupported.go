package documentshttp

import "net/http"

func (handler *Handler) legacyFolderArchive(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "folder_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.recordLegacyHit(w, r, permissionFolderRead, operationLegacyArchive, id)
}

func (handler *Handler) legacyCreateDocument(w http.ResponseWriter, r *http.Request) {
	handler.recordLegacyHit(w, r, permissionItemCreate, operationLegacyUpload, "")
}

func (handler *Handler) legacyCreateVersion(w http.ResponseWriter, r *http.Request) {
	id, valid := pathID(r, "document_id")
	if !valid {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.recordLegacyHit(w, r, permissionItemVersion, operationLegacyUpload, id)
}

func (handler *Handler) recordLegacyHit(w http.ResponseWriter, r *http.Request,
	permission, operation, resourceID string,
) {
	access, err := handler.accessFor(r, permission, operation)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	_, err = handler.application.Dispatch(r.Context(), Request{Access: access,
		Operation: operation, ResourceID: resourceID})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeFailure(w, r, ErrClientUpgradeRequired)
}
