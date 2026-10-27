package systemtransfer

import (
	"time"

	"github.com/google/uuid"
)

type ExportWrite struct {
	Evidence                            ControlEvidence
	ExportID, JobID, ReceiptID          uuid.UUID
	CatalogVersionID                    uuid.UUID
	CatalogSHA256, QuerySHA256          [32]byte
	RequestSHA256, IdempotencyKeySHA256 [32]byte
	QueryDocument                       []byte
	Format                              string
	CreatedAt                           time.Time
}

func (value ExportWrite) Validate() error {
	if value.Evidence.Validate() != nil || !controlV7(value.ExportID) ||
		!controlV7(value.JobID) || !controlV7(value.ReceiptID) ||
		!controlV7(value.CatalogVersionID) || value.CatalogSHA256 == [32]byte{} ||
		value.QuerySHA256 == [32]byte{} || value.RequestSHA256 == [32]byte{} ||
		value.IdempotencyKeySHA256 == [32]byte{} || len(value.QueryDocument) < 2 ||
		len(value.QueryDocument) > 262144 ||
		(value.Format != "csv" && value.Format != "jsonl" && value.Format != "xlsx") ||
		!controlTime(value.CreatedAt) {
		return ErrControlInvalid
	}
	return nil
}

type ExportRecord struct {
	ExportID, JobID uuid.UUID
	Resource        string
	Format, State   string
	Rows, Bytes     int64
	CreatedAt       time.Time
	ReadyAt         *time.Time
	ExpiresAt       *time.Time
	Replay          bool
}

type ExportArtifact struct {
	ExportRecord
	ObjectKey, ObjectVersion, KeyID string
	HeaderSHA256, CiphertextSHA256  [32]byte
	CiphertextBytes                 int64
	WrappedDEK                      []byte
	WrapNonce                       [12]byte
	AttemptGeneration               int64
}
