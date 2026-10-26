package iam

import "net/http"

func (handler *SystemTransferHTTP) reportSystemImport(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	importID, idOK := systemTransferV7Path(r, "import_id")
	after, limit, pageOK := systemTransferReportPage(r)
	if !resourceOK || !idOK || !pageOK {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	access, err := handler.transferAccess(r, resource, "import", SystemImportReportOperation)
	if err == nil {
		view, reportErr := handler.application.ReportSystemImport(r.Context(),
			SystemImportReportInput{SystemImportReadInput: SystemImportReadInput{
				Access: access, Resource: resource, ImportID: importID},
				AfterRow: after, Limit: limit})
		err = reportErr
		if err == nil {
			handler.writeTransferJSON(w, r, http.StatusOK, view)
			return
		}
	}
	handler.writeTransferFailure(w, r, err)
}
