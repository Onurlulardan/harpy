package systemtransfer

import (
	"context"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

type controlStoreFixture struct {
	cancelExport       func(ControlRead, time.Time) (ExportRecord, error)
	cancelImport       func(ControlRead, time.Time) (ImportRecord, error)
	readExport         func(ControlRead) (ExportRecord, error)
	readExportArtifact func(ControlRead, time.Time) (ExportArtifact, error)
	readImport         func(ControlRead) (ImportRecord, error)
	readReceipts       func(ControlRead, int64, int) (ReceiptPage, error)
	submitExport       func(ExportWrite) (ExportRecord, error)
	submitImport       func(ImportWrite) (ImportRecord, error)
}

func (*controlStoreFixture) RegisterSource(context.Context, SourceWrite) (
	SourceRecord, error) {
	return SourceRecord{}, ErrControlUnavailable
}

func (store *controlStoreFixture) ExportRequest(_ context.Context,
	write ExportWrite) (ExportRecord, error) {
	return store.submitExport(write)
}

func (store *controlStoreFixture) GetExport(_ context.Context,
	read ControlRead) (ExportRecord, error) {
	return store.readExport(read)
}

func (store *controlStoreFixture) DownloadExportArtifact(_ context.Context,
	read ControlRead, now time.Time) (ExportArtifact, error) {
	return store.readExportArtifact(read, now)
}

func (store *controlStoreFixture) CancelExport(_ context.Context,
	read ControlRead, now time.Time) (ExportRecord, error) {
	return store.cancelExport(read, now)
}

func (store *controlStoreFixture) ImportRequest(_ context.Context,
	write ImportWrite) (ImportRecord, error) {
	return store.submitImport(write)
}

func (store *controlStoreFixture) ResolveImport(_ context.Context,
	read ControlRead) (ImportRecord, error) {
	return store.readImport(read)
}

func (store *controlStoreFixture) DeleteImport(_ context.Context,
	read ControlRead, now time.Time) (ImportRecord, error) {
	return store.cancelImport(read, now)
}

func (store *controlStoreFixture) ListReceipts(_ context.Context,
	read ControlRead, after int64, limit int) (ReceiptPage, error) {
	return store.readReceipts(read, after, limit)
}

func controlFixtureRead() ControlRead {
	return ControlRead{Evidence: ControlEvidence{
		TenantID:     uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		ActorID:      uuid.MustParse("91000000-0000-4000-8000-000000000001"),
		TokenVersion: 3, AuthorizationVersion: 5, AuthorizationClock: 7,
		Resource: publicresource.Users,
	}, TransferID: uuid.MustParse("019be000-0000-7000-8000-000000000001")}
}

func controlFixtureTime() time.Time {
	return time.Date(2026, 7, 18, 10, 0, 0, 0, time.UTC)
}

func controlFixtureExportWrite() ExportWrite {
	read := controlFixtureRead()
	return ExportWrite{Evidence: read.Evidence, ExportID: read.TransferID,
		JobID:            uuid.MustParse("019be000-0000-7000-8000-000000000002"),
		ReceiptID:        uuid.MustParse("019be000-0000-7000-8000-000000000003"),
		CatalogVersionID: uuid.MustParse("019be000-0000-7000-8000-000000000004"),
		CatalogSHA256:    [32]byte{1}, QuerySHA256: [32]byte{2},
		RequestSHA256: [32]byte{3}, IdempotencyKeySHA256: [32]byte{4},
		QueryDocument: []byte("{}"), Format: "csv", CreatedAt: controlFixtureTime()}
}

func controlFixtureImportWrite() ImportWrite {
	read := controlFixtureRead()
	return ImportWrite{Evidence: read.Evidence, ImportID: read.TransferID,
		JobID:            uuid.MustParse("019be000-0000-7000-8000-000000000005"),
		ReceiptID:        uuid.MustParse("019be000-0000-7000-8000-000000000006"),
		SourceID:         uuid.MustParse("019be000-0000-7000-8000-000000000007"),
		CatalogVersionID: uuid.MustParse("019be000-0000-7000-8000-000000000008"),
		CatalogSHA256:    [32]byte{1}, IdentityManifestSHA256: [32]byte{2},
		RequestSHA256: [32]byte{3}, IdempotencyKeySHA256: [32]byte{4},
		MappingSHA256: [32]byte{5}, DuplicateAction: "skip",
		MappingDocument: []byte("{}"), FieldCount: 1, CreatedAt: controlFixtureTime()}
}
