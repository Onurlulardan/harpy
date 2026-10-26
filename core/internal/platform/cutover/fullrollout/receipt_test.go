package fullrollout

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type receiptRecorder struct {
	digest [sha256.Size]byte
	value  PersistedCompletion
	err    error
}

func (recorder *receiptRecorder) Record(
	_ context.Context, value PersistedCompletion, digest [sha256.Size]byte,
) ([sha256.Size]byte, error) {
	recorder.value = value
	if recorder.err != nil {
		return [sha256.Size]byte{}, recorder.err
	}
	if recorder.digest != [sha256.Size]byte{} {
		return recorder.digest, nil
	}
	return digest, nil
}

func TestRecordPersistsOneBoundCompletionDigest(t *testing.T) {
	evidence := validCompletion(t)
	recorder := &receiptRecorder{}
	value := PersistedCompletion{
		ReceiptID:     uuid.MustParse("30000000-0000-4000-8000-000000000001"),
		EnvironmentID: uuid.UUID(evidence.Steps[0].ScopeID),
		Evidence:      evidence,
		RecordedAt:    evidence.Now,
	}
	receipt, err := Record(context.Background(), recorder, value)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := receipt.SHA256()
	if err != nil || digest == [sha256.Size]byte{} || recorder.value.ReceiptID != value.ReceiptID {
		t.Fatalf("receipt was not bound: %x %v", digest, err)
	}

	tampered := value
	tampered.Evidence.V1RemovalSHA256[0] ^= 0xff
	if digestCompletion(tampered) == digest {
		t.Fatal("V1 removal tamper retained completion digest")
	}
}

func TestRecordRejectsDatabaseDigestMismatch(t *testing.T) {
	evidence := validCompletion(t)
	value := PersistedCompletion{
		ReceiptID:     uuid.MustParse("30000000-0000-4000-8000-000000000002"),
		EnvironmentID: uuid.UUID(evidence.Steps[0].ScopeID),
		Evidence:      evidence,
		RecordedAt:    evidence.Now,
	}
	recorder := &receiptRecorder{digest: sha256.Sum256([]byte("different"))}
	if _, err := Record(context.Background(), recorder, value); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("digest mismatch error = %v", err)
	}
}

func TestRecordRejectsScopeOrEvidenceTimeMismatch(t *testing.T) {
	value := validPersistedCompletion(t)
	value.RecordedAt = value.RecordedAt.Add(time.Microsecond)
	if _, err := Record(context.Background(), &receiptRecorder{}, value); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("evidence time mismatch error = %v", err)
	}

	value = validPersistedCompletion(t)
	tenantID := value.Evidence.Steps[1].ScopeID
	value.EnvironmentID = uuid.UUID(tenantID)
	value.Evidence.Steps[0].ScopeID = tenantID
	if _, err := Record(context.Background(), &receiptRecorder{}, value); !errors.Is(err, ErrInvalidEvidence) {
		t.Fatalf("environment and tenant collision error = %v", err)
	}
}
