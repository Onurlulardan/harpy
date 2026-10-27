package iam

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

const SystemTransferMaximumUploadBytes int64 = 512 << 20

type SystemExportSubmitInput struct {
	Access         SystemTransferAccess
	Resource       SystemTransferResource
	Format         string
	QueryDocument  json.RawMessage
	IdempotencyKey string
}

type SystemExportReadInput struct {
	Access   SystemTransferAccess
	Resource SystemTransferResource
	ExportID uuid.UUID
}

type SystemImportSourceInput struct {
	Access         SystemTransferAccess
	Resource       SystemTransferResource
	Format         string
	Body           io.Reader
	IdempotencyKey string
}

type SystemImportSubmitInput struct {
	Access          SystemTransferAccess
	Resource        SystemTransferResource
	SourceID        uuid.UUID
	DuplicateAction string
	MappingDocument json.RawMessage
	IdempotencyKey  string
}

type SystemImportReadInput struct {
	Access   SystemTransferAccess
	Resource SystemTransferResource
	ImportID uuid.UUID
}

type SystemImportReportInput struct {
	SystemImportReadInput
	AfterRow int64
	Limit    int
}
