package documentshttp

import (
	"encoding/json"
	"net/http"
)

func (handler *Handler) dispatchJSON(w http.ResponseWriter, r *http.Request,
	permission, operation, resourceID, secondaryID string, document json.RawMessage,
) {
	access, err := handler.accessFor(r, permission, operation)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	key := ""
	if operationNeedsIdempotency(operation) {
		var valid bool
		key, valid = idempotencyKey(r)
		if !valid {
			handler.writeFailure(w, r, errRequestInvalid)
			return
		}
	}
	response, err := handler.application.Dispatch(r.Context(), Request{Access: access,
		Operation: operation, ResourceID: resourceID, SecondaryID: secondaryID,
		IdempotencyKey: key, Query: queryCopy(r), Document: document})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.writeResponse(w, r, response)
}

func operationNeedsIdempotency(operation string) bool {
	switch operation {
	case operationFolderCreate, operationFolderUpdate, operationFolderMove,
		operationFolderDelete, operationFolderRestore, operationItemUpdate,
		operationItemMove, operationItemCopy, operationItemDelete, operationItemRestore,
		operationItemPurge, operationItemCancel, operationUploadCreate,
		operationUploadCancel, operationArchiveCreate:
		return true
	default:
		return false
	}
}

func (handler *Handler) dispatchRead(w http.ResponseWriter, r *http.Request,
	permission, operation, resourceID, secondaryID string,
) {
	if !bodyAbsent(r) {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	handler.dispatchJSON(w, r, permission, operation, resourceID, secondaryID, nil)
}

func (handler *Handler) dispatchDocument(w http.ResponseWriter, r *http.Request,
	permission, operation, resourceID, secondaryID string,
) {
	if r == nil || r.URL == nil || r.URL.RawQuery != "" {
		handler.writeFailure(w, r, errRequestInvalid)
		return
	}
	document, err := decodeDocument(w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	defer clear(document)
	handler.dispatchJSON(w, r, permission, operation, resourceID, secondaryID, document)
}
