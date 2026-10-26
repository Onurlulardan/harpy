package iam

import "context"

type SystemTransferApplication interface {
	SubmitSystemExport(context.Context, SystemExportSubmitInput) (SystemExportView, error)
	GetSystemExport(context.Context, SystemExportReadInput) (SystemExportView, error)
	CancelSystemExport(context.Context, SystemExportReadInput) (SystemExportView, error)
	DownloadSystemExport(context.Context, SystemExportReadInput) (SystemExportContent, error)
	CreateSystemImportSource(context.Context, SystemImportSourceInput) (
		SystemImportSourceView, error,
	)
	SubmitSystemImport(context.Context, SystemImportSubmitInput) (SystemImportView, error)
	GetSystemImport(context.Context, SystemImportReadInput) (SystemImportView, error)
	CancelSystemImport(context.Context, SystemImportReadInput) (SystemImportView, error)
	ReportSystemImport(context.Context, SystemImportReportInput) (SystemImportReportView, error)
}
