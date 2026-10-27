package dashboardexporthttp

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func (handler *Handler) Content(writer http.ResponseWriter, request *http.Request) {
	access, err := handler.runtime.AccessFor(request,
		httptransport.PermissionDashboardExport, httptransport.OperationDashboardExportContent)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	dashboardID, dashboardOK := httptransport.PathV7(request, "dashboard_id")
	exportID, exportOK := httptransport.PathV7(request, "export_id")
	if !dashboardOK || !exportOK || request.URL.RawQuery != "" || !httptransport.BodyAbsent(request) {
		handler.runtime.Failure(writer, request, httptransport.ErrInvalid)
		return
	}
	content, err := handler.application.OpenDashboardExportContent(
		request.Context(), access, dashboardID, exportID)
	if err != nil {
		handler.runtime.Failure(writer, request, err)
		return
	}
	defer func() {
		if closeErr := content.Reader.Close(); closeErr != nil {
			handler.runtime.Report(errors.New("analytics export content close failed"))
		}
	}()
	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": content.Filename,
	})
	if disposition == "" || content.Length < 1 || content.ContentType == "" {
		handler.runtime.Failure(writer, request, httptransport.ErrUnavailable)
		return
	}
	header := writer.Header()
	header.Set("Content-Type", content.ContentType)
	header.Set("Content-Disposition", disposition)
	header.Set("Content-Length", strconv.FormatInt(content.Length, 10))
	header.Set("Cache-Control", "no-store")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Correlation-ID", handler.runtime.Correlation(request))
	writer.WriteHeader(http.StatusOK)
	written, copyErr := io.CopyN(writer, content.Reader, content.Length)
	if copyErr != nil || written != content.Length {
		handler.runtime.Report(errors.New("analytics export content write failed"))
	}
}
