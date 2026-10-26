package iam

import "net/http"

func (handler *SystemTransferHTTP) getSystemImport(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	importID, idOK := systemTransferV7Path(r, "import_id")
	if !resourceOK || !idOK || !systemTransferQueryAbsent(r) {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	access, err := handler.transferAccess(r, resource, "import", SystemImportGetOperation)
	if err == nil {
		view, getErr := handler.application.GetSystemImport(r.Context(),
			SystemImportReadInput{Access: access, Resource: resource, ImportID: importID})
		err = getErr
		if err == nil {
			handler.writeTransferJSON(w, r, http.StatusOK, systemImportHTTPView(view))
			return
		}
	}
	handler.writeTransferFailure(w, r, err)
}
