package systemtransfer

import (
	"time"

	"github.com/google/uuid"
)

type ImportWrite struct {
	Evidence                              ControlEvidence
	ImportID, JobID, ReceiptID, SourceID  uuid.UUID
	CatalogVersionID                      uuid.UUID
	CatalogSHA256, IdentityManifestSHA256 [32]byte
	RequestSHA256, IdempotencyKeySHA256   [32]byte
	MappingSHA256                         [32]byte
	DuplicateAction                       string
	MappingDocument                       []byte
	FieldCount                            int
	CreatedAt                             time.Time
}

func (value ImportWrite) Validate() error {
	if value.Evidence.Validate() != nil || !controlV7(value.ImportID) ||
		!controlV7(value.JobID) || !controlV7(value.ReceiptID) ||
		!controlV7(value.SourceID) || !controlV7(value.CatalogVersionID) ||
		value.CatalogSHA256 == [32]byte{} || value.IdentityManifestSHA256 == [32]byte{} ||
		value.RequestSHA256 == [32]byte{} || value.IdempotencyKeySHA256 == [32]byte{} ||
		value.MappingSHA256 == [32]byte{} ||
		(value.DuplicateAction != "skip" && value.DuplicateAction != "update") ||
		len(value.MappingDocument) < 2 || len(value.MappingDocument) > 1048576 ||
		value.FieldCount < 1 || value.FieldCount > 500 || !controlTime(value.CreatedAt) {
		return ErrControlInvalid
	}
	return nil
}

type ImportRecord struct {
	ImportID, JobID            uuid.UUID
	Resource, State            string
	RowsSeen, Created, Updated int64
	Skipped, Failed            int64
	CreatedAt                  time.Time
	CompletedAt                *time.Time
	Replay                     bool
}

type ReceiptRecord struct {
	RowIndex                      int64
	Outcome, OwnerKind, ErrorCode string
	OwnerID                       *uuid.UUID
	SourceSHA256, ResultSHA256    [32]byte
}

type ReceiptPage struct {
	Rows      []ReceiptRecord
	NextAfter int64
	HasMore   bool
}
