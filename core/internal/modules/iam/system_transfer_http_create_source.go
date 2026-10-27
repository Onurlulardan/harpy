package iam

import "net/http"

func (handler *SystemTransferHTTP) createSystemImportSource(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	key, keyOK := systemTransferIdempotencyKey(r)
	if !resourceOK || !keyOK || !systemTransferQueryAbsent(r) || r.Body == nil {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	format, err := systemTransferImportFormat(r)
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	if r.ContentLength > SystemTransferMaximumUploadBytes {
		handler.writeTransferFailure(w, r, ErrSystemTransferTooLarge)
		return
	}
	access, err := handler.transferAccess(r, resource, "import", SystemImportSourceOperation)
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, SystemTransferMaximumUploadBytes)
	view, err := handler.application.CreateSystemImportSource(r.Context(),
		SystemImportSourceInput{Access: access, Resource: resource,
			Format: format, Body: r.Body, IdempotencyKey: key})
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	w.Header().Set("Location", systemImportSourceLocation(resource, view.SourceID.String()))
	status := http.StatusCreated
	if view.Replay {
		status = http.StatusOK
	}
	handler.writeTransferJSON(w, r, status, systemImportSourceHTTPView(view))
}
