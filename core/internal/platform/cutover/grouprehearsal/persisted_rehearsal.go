package grouprehearsal

import (
	"context"
	"crypto/sha256"
	"math"
	"time"

	"github.com/google/uuid"
)

type PersistedRehearsal struct {
	ReceiptID  uuid.UUID
	Evidence   GroupEvidence
	RecordedAt time.Time
}

type Recorder interface {
	Record(context.Context, PersistedRehearsal, [sha256.Size]byte) (
		[sha256.Size]byte, error,
	)
}

func validatePersisted(value PersistedRehearsal) error {
	if value.ReceiptID == uuid.Nil || value.RecordedAt.IsZero() ||
		value.RecordedAt.Location() != time.UTC ||
		value.RecordedAt.Nanosecond()%1_000 != 0 {
		return ErrInvalid
	}
	evidence := value.Evidence
	if evidence.TenantID == uuid.Nil || evidence.Generation < 1 ||
		evidence.GroupID != CanonicalGroupID || len(evidence.Members) != 2 ||
		!evidence.QuiesceDrained || !evidence.UsageFlushed ||
		evidence.ProductionDualWrite {
		return ErrInvalid
	}
	want := []string{"billing", "forms"}
	seen := map[uuid.UUID]struct{}{value.ReceiptID: {}}
	for index, member := range evidence.Members {
		if member.Domain != want[index] || member.OperationID == uuid.Nil ||
			member.RunID == uuid.Nil || member.RollbackReceiptID == uuid.Nil ||
			member.SourceCount > math.MaxInt64 || member.TargetCount > math.MaxInt64 ||
			member.DifferenceCount != 0 || member.SourceCount != member.TargetCount ||
			member.AcceptanceSHA256 == [sha256.Size]byte{} ||
			member.RollbackSHA256 == [sha256.Size]byte{} ||
			member.WriterBefore != WriterV1 || member.WriterAfter != WriterV1 {
			return ErrInvalid
		}
		for _, identity := range []uuid.UUID{
			member.OperationID, member.RunID, member.RollbackReceiptID,
		} {
			if _, duplicate := seen[identity]; duplicate {
				return ErrInvalid
			}
			seen[identity] = struct{}{}
		}
	}
	return nil
}
