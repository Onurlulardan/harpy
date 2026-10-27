package iam

import "net/http"

func (handler *SystemTransferHTTP) submitSystemExport(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	key, keyOK := systemTransferIdempotencyKey(r)
	if !resourceOK || !keyOK || !systemTransferQueryAbsent(r) {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	access, err := handler.transferAccess(r, resource, "export", SystemExportSubmitOperation)
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	document, err := decodeSystemTransferJSON[systemExportSubmitDocument](w, r)
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	defer clear(document.Query)
	if !systemTransferExportFormat(document.Format) || len(document.Query) < 2 {
		handler.writeTransferFailure(w, r, ErrSystemTransferInvalid)
		return
	}
	view, err := handler.application.SubmitSystemExport(r.Context(),
		SystemExportSubmitInput{Access: access, Resource: resource,
			Format: document.Format, QueryDocument: document.Query, IdempotencyKey: key})
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	w.Header().Set("Location", systemExportLocation(resource, view.ExportID.String()))
	handler.writeTransferJSON(w, r, http.StatusAccepted, systemExportHTTPView(view))
}

func systemTransferExportFormat(value string) bool {
	return value == "csv" || value == "jsonl" || value == "xlsx"
}
