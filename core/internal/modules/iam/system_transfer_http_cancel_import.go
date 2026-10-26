package iam

import "net/http"

func (handler *SystemTransferHTTP) cancelSystemImport(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	importID, idOK := systemTransferV7Path(r, "import_id")
	if !resourceOK || !idOK || !systemTransferQueryAbsent(r) {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	access, err := handler.transferAccess(r, resource, "import", SystemImportCancelOperation)
	if err == nil {
		view, cancelErr := handler.application.CancelSystemImport(r.Context(),
			SystemImportReadInput{Access: access, Resource: resource, ImportID: importID})
		err = cancelErr
		if err == nil {
			handler.writeTransferJSON(w, r, http.StatusOK, systemImportHTTPView(view))
			return
		}
	}
	handler.writeTransferFailure(w, r, err)
}
