package iam

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type SystemExportView struct {
	ExportID  uuid.UUID
	Resource  SystemTransferResource
	Format    string
	State     string
	Rows      int64
	Bytes     int64
	CreatedAt time.Time
	ReadyAt   *time.Time
	ExpiresAt *time.Time
	Replay    bool
}

type SystemExportContent struct {
	Body              io.ReadCloser
	Size              int64
	Name, ContentType string
}

type SystemImportSourceView struct {
	SourceID   uuid.UUID
	Resource   SystemTransferResource
	Format     string
	Bytes      int64
	SHA256     [32]byte
	UploadedAt time.Time
	Replay     bool
	Columns    []SystemImportSourceColumnView
	Samples    []map[string]any
}

type SystemImportSourceColumnView struct {
	Name string
	Kind string
}

type SystemImportView struct {
	ImportID                                    uuid.UUID
	Resource                                    SystemTransferResource
	State                                       string
	RowsSeen, Created, Updated, Skipped, Failed int64
	CreatedAt                                   time.Time
	CompletedAt                                 *time.Time
	Replay                                      bool
}

type SystemImportReceiptView struct {
	RowIndex                                                           int64
	Outcome, OwnerKind, OwnerID, ErrorCode, SourceDigest, ResultDigest string
}

type SystemImportReportView struct {
	ImportID     uuid.UUID
	Receipts     []SystemImportReceiptView
	NextAfterRow int64
	HasMore      bool
}
