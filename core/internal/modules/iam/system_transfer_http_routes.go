package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

const systemTransferRoot = "/api/iam/system-resources/{resource}"
const systemExportCollection = systemTransferRoot + "/record-exports"
const systemExportMember = systemExportCollection + "/{export_id}"
const systemImportSources = systemTransferRoot + "/record-import-sources"
const systemImportCollection = systemTransferRoot + "/record-imports"
const systemImportMember = systemImportCollection + "/{import_id}"

func (handler *SystemTransferHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	routes := []httpedge.Route{
		{Method: http.MethodPost, Pattern: systemExportCollection,
			Handler: http.HandlerFunc(handler.submitSystemExport)},
		{Method: http.MethodGet, Pattern: systemExportMember,
			Handler: http.HandlerFunc(handler.getSystemExport)},
		{Method: http.MethodDelete, Pattern: systemExportMember,
			Handler: http.HandlerFunc(handler.cancelSystemExport)},
		{Method: http.MethodGet, Pattern: systemExportMember + "/content",
			Handler: http.HandlerFunc(handler.downloadSystemExport)},
		{Method: http.MethodPost, Pattern: systemImportSources,
			Handler: http.HandlerFunc(handler.createSystemImportSource)},
		{Method: http.MethodPost, Pattern: systemImportCollection,
			Handler: http.HandlerFunc(handler.submitSystemImport)},
		{Method: http.MethodGet, Pattern: systemImportMember,
			Handler: http.HandlerFunc(handler.getSystemImport)},
		{Method: http.MethodDelete, Pattern: systemImportMember,
			Handler: http.HandlerFunc(handler.cancelSystemImport)},
		{Method: http.MethodGet, Pattern: systemImportMember + "/report",
			Handler: http.HandlerFunc(handler.reportSystemImport)},
	}
	return append(routes, handler.systemTransferV1Routes()...)
}
