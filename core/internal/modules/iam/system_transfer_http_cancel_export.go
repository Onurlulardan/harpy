package iam

import "net/http"

func (handler *SystemTransferHTTP) cancelSystemExport(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	exportID, idOK := systemTransferV7Path(r, "export_id")
	if !resourceOK || !idOK || !systemTransferQueryAbsent(r) {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	access, err := handler.transferAccess(r, resource, "export", SystemExportCancelOperation)
	if err == nil {
		view, cancelErr := handler.application.CancelSystemExport(r.Context(),
			SystemExportReadInput{Access: access, Resource: resource, ExportID: exportID})
		err = cancelErr
		if err == nil {
			handler.writeTransferJSON(w, r, http.StatusOK, systemExportHTTPView(view))
			return
		}
	}
	handler.writeTransferFailure(w, r, err)
}
