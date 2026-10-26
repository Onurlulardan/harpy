package link

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type FormRecordAdoptionCheckpoint struct {
	OperationID, DefinitionID, Cursor domain.ID
	Epoch                             int64
	RecordCount, LinkCount            int64
	SourceSHA256, LinkSHA256          [32]byte
	Done                              bool
}

type FormRecordAdoptionSeal struct {
	Checkpoint          FormRecordAdoptionCheckpoint
	SourceReceiptSHA256 [32]byte
	IssuedAt            time.Time
}

type FormRecordAdoptionReceipt struct {
	SourceReceiptSHA256, ReceiptSHA256 [32]byte
	IssuedAt                           time.Time
}
