package iam

import "net/http"

func (handler *SystemTransferHTTP) submitSystemImport(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	key, keyOK := systemTransferIdempotencyKey(r)
	if !resourceOK || !keyOK || !systemTransferQueryAbsent(r) {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	access, err := handler.transferAccess(r, resource, "import", SystemImportSubmitOperation)
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	document, err := decodeSystemTransferJSON[systemImportSubmitDocument](w, r)
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	defer clear(document.Mapping)
	sourceID, parseErr := uuidV7(document.SourceID)
	if parseErr != nil || document.DuplicateAction != "skip" &&
		document.DuplicateAction != "update" || len(document.Mapping) < 2 {
		handler.writeTransferFailure(w, r, ErrSystemTransferInvalid)
		return
	}
	view, err := handler.application.SubmitSystemImport(r.Context(),
		SystemImportSubmitInput{Access: access, Resource: resource, SourceID: sourceID,
			DuplicateAction: document.DuplicateAction, MappingDocument: document.Mapping,
			IdempotencyKey: key})
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	w.Header().Set("Location", systemImportLocation(resource, view.ImportID.String()))
	handler.writeTransferJSON(w, r, http.StatusAccepted, systemImportHTTPView(view))
}
