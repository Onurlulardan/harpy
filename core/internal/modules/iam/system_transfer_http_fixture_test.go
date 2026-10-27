package iam

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type systemTransferAccessStub struct {
	request SystemTransferAccessRequest
}

func (stub *systemTransferAccessStub) VerifySystemTransfer(_ context.Context,
	request SystemTransferAccessRequest) (SystemTransferAccess, error) {
	stub.request = request
	return NewSystemTransferAccess(
		uuid.MustParse("019be000-0000-7000-8000-000000000101"),
		uuid.MustParse("019be000-0000-7000-8000-000000000102"),
		1, 2, 3, request.Resource, request.Permission, request.Operation,
	)
}

type systemTransferApplicationStub struct {
	exportInput SystemExportSubmitInput
	sourceInput SystemImportSourceInput
}

func (stub *systemTransferApplicationStub) SubmitSystemExport(_ context.Context,
	input SystemExportSubmitInput) (SystemExportView, error) {
	stub.exportInput = input
	return SystemExportView{ExportID: uuid.MustParse(
		"019be000-0000-7000-8000-000000000201"), Resource: input.Resource,
		Format: input.Format, State: "waiting", CreatedAt: time.Unix(1, 0).UTC()}, nil
}

func (*systemTransferApplicationStub) GetSystemExport(context.Context,
	SystemExportReadInput) (SystemExportView, error) {
	return SystemExportView{}, ErrSystemTransferNotFound
}
func (*systemTransferApplicationStub) CancelSystemExport(context.Context,
	SystemExportReadInput) (SystemExportView, error) {
	return SystemExportView{}, ErrSystemTransferNotFound
}
func (*systemTransferApplicationStub) DownloadSystemExport(context.Context,
	SystemExportReadInput) (SystemExportContent, error) {
	return SystemExportContent{}, ErrSystemTransferNotFound
}
func (stub *systemTransferApplicationStub) CreateSystemImportSource(_ context.Context,
	input SystemImportSourceInput) (SystemImportSourceView, error) {
	stub.sourceInput = input
	return SystemImportSourceView{SourceID: uuid.MustParse(
		"019be000-0000-7000-8000-000000000202"), Resource: input.Resource,
		Format: input.Format, Bytes: 1, UploadedAt: time.Unix(1, 0).UTC()}, nil
}
func (*systemTransferApplicationStub) SubmitSystemImport(context.Context,
	SystemImportSubmitInput) (SystemImportView, error) {
	return SystemImportView{}, ErrSystemTransferNotFound
}
func (*systemTransferApplicationStub) GetSystemImport(context.Context,
	SystemImportReadInput) (SystemImportView, error) {
	return SystemImportView{}, ErrSystemTransferNotFound
}
func (*systemTransferApplicationStub) CancelSystemImport(context.Context,
	SystemImportReadInput) (SystemImportView, error) {
	return SystemImportView{}, ErrSystemTransferNotFound
}
func (*systemTransferApplicationStub) ReportSystemImport(context.Context,
	SystemImportReportInput) (SystemImportReportView, error) {
	return SystemImportReportView{}, ErrSystemTransferNotFound
}

func newSystemTransferHTTPFixture(t *testing.T) (*SystemTransferHTTP,
	*systemTransferAccessStub, *systemTransferApplicationStub) {
	t.Helper()
	access := new(systemTransferAccessStub)
	application := new(systemTransferApplicationStub)
	handler, err := NewSystemTransferHTTP(SystemTransferHTTPDependencies{
		Access: access, Application: application, Report: func(error) {},
	})
	if err != nil {
		t.Fatal(err)
	}
	return handler, access, application
}
