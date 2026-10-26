package iam

import "net/http"

func (handler *SystemTransferHTTP) getSystemExport(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	exportID, idOK := systemTransferV7Path(r, "export_id")
	if !resourceOK || !idOK || !systemTransferQueryAbsent(r) {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	access, err := handler.transferAccess(r, resource, "export", SystemExportGetOperation)
	if err == nil {
		view, getErr := handler.application.GetSystemExport(r.Context(),
			SystemExportReadInput{Access: access, Resource: resource, ExportID: exportID})
		err = getErr
		if err == nil {
			handler.writeTransferJSON(w, r, http.StatusOK, systemExportHTTPView(view))
			return
		}
	}
	handler.writeTransferFailure(w, r, err)
}
