package iam

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (handler *SystemTransferHTTP) downloadSystemExport(w http.ResponseWriter,
	r *http.Request) {
	resource, resourceOK := systemTransferPath(r)
	exportID, idOK := systemTransferV7Path(r, "export_id")
	if !resourceOK || !idOK || !systemTransferQueryAbsent(r) {
		handler.writeTransferFailure(w, r, errSystemTransferRequest)
		return
	}
	access, err := handler.transferAccess(r, resource, "export", SystemExportDownloadOperation)
	if err != nil {
		handler.writeTransferFailure(w, r, err)
		return
	}
	content, err := handler.application.DownloadSystemExport(r.Context(),
		SystemExportReadInput{Access: access, Resource: resource, ExportID: exportID})
	if err != nil || content.Body == nil || content.Size < 0 ||
		!systemExportContentTypeAllowed(content.ContentType) {
		if err == nil {
			err = ErrSystemTransferUnavailable
		}
		handler.writeTransferFailure(w, r, err)
		return
	}
	defer func() {
		if closeErr := content.Body.Close(); closeErr != nil {
			handler.report(errors.New("IAM export content close failed"))
		}
	}()
	w.Header().Set("Content-Type", content.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s"`,
		resource, content.Name))
	w.Header().Set("Content-Length", fmt.Sprint(content.Size))
	w.WriteHeader(http.StatusOK)
	if _, err = io.CopyN(w, content.Body, content.Size); err != nil {
		handler.report(errors.New("IAM export content write failed"))
	}
}

func systemExportContentTypeAllowed(value string) bool {
	return value == "text/csv; charset=utf-8" || value == "application/x-ndjson" ||
		value == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}
