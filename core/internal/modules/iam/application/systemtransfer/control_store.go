package systemtransfer

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ControlScopeAccess uint8

const (
	ControlReadAccess ControlScopeAccess = iota + 1
	ControlCommandAccess
)

type ControlStore interface {
	RegisterSource(context.Context, SourceWrite) (SourceRecord, error)
	ExportRequest(context.Context, ExportWrite) (ExportRecord, error)
	GetExport(context.Context, ControlRead) (ExportRecord, error)
	DownloadExportArtifact(context.Context, ControlRead, time.Time) (ExportArtifact, error)
	CancelExport(context.Context, ControlRead, time.Time) (ExportRecord, error)
	ImportRequest(context.Context, ImportWrite) (ImportRecord, error)
	ResolveImport(context.Context, ControlRead) (ImportRecord, error)
	DeleteImport(context.Context, ControlRead, time.Time) (ImportRecord, error)
	ListReceipts(context.Context, ControlRead, int64, int) (ReceiptPage, error)
}

type ControlStoreBinder interface {
	BindControl(context.Context, platformtx.TenantScope, ControlScopeAccess) (
		ControlStore, error)
}

func controlV7(value [16]byte) bool {
	return value != [16]byte{} && value[6]>>4 == 7 && value[8]&0xc0 == 0x80
}

func controlTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}
