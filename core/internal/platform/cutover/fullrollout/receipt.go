package fullrollout

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PersistedCompletion struct {
	ReceiptID     uuid.UUID
	EnvironmentID uuid.UUID
	Evidence      CompletionEvidence
	RecordedAt    time.Time
}

type Recorder interface {
	Record(context.Context, PersistedCompletion, [sha256.Size]byte) (
		[sha256.Size]byte, error,
	)
}

type Receipt struct{ digest [sha256.Size]byte }

func Record(
	ctx context.Context, recorder Recorder, value PersistedCompletion,
) (Receipt, error) {
	tenantID, stepErr := validateSteps(value.Evidence.Steps)
	if ctx == nil || ctx.Err() != nil || recorder == nil ||
		value.ReceiptID == uuid.Nil || value.EnvironmentID == uuid.Nil ||
		!validEvidenceTime(value.RecordedAt) ||
		stepErr != nil || uuid.UUID(tenantID) == value.EnvironmentID ||
		!value.RecordedAt.Equal(value.Evidence.Now) ||
		EvaluateCompletion(value.Evidence) != nil ||
		len(value.Evidence.Steps) == 0 ||
		uuid.UUID(value.Evidence.Steps[0].ScopeID) != value.EnvironmentID {
		return Receipt{}, ErrInvalidEvidence
	}
	digest := digestCompletion(value)
	persisted, err := recorder.Record(ctx, value, digest)
	if err != nil {
		return Receipt{}, fmt.Errorf("record full rollout completion: %w", err)
	}
	if persisted != digest {
		return Receipt{}, ErrInvalidEvidence
	}
	return Receipt{digest: digest}, nil
}

func (value Receipt) SHA256() ([sha256.Size]byte, error) {
	if value.digest == [sha256.Size]byte{} {
		return [sha256.Size]byte{}, ErrInvalidEvidence
	}
	return value.digest, nil
}

func validEvidenceTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func (Receipt) String() string   { return "[REDACTED]" }
func (Receipt) GoString() string { return "[REDACTED]" }
