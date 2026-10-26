package grouprehearsal

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memoryRecorder struct {
	digest [sha256.Size]byte
	err    error
}

func (recorder *memoryRecorder) Record(
	_ context.Context, _ PersistedRehearsal, digest [sha256.Size]byte,
) ([sha256.Size]byte, error) {
	if recorder.err != nil {
		return [sha256.Size]byte{}, recorder.err
	}
	recorder.digest = digest
	return digest, nil
}

func TestRecordRequiresCanonicalPersistedBillingFormsEvidence(t *testing.T) {
	value := validPersistedRehearsal()
	recorder := &memoryRecorder{}
	accepted, err := Record(context.Background(), recorder, value)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := accepted.SHA256()
	if err != nil || digest == [sha256.Size]byte{} || digest != recorder.digest {
		t.Fatalf("persisted digest=%x err=%v", digest, err)
	}
	value.Evidence.GroupID = MandatoryGroupID
	if _, err := Record(context.Background(), recorder, value); err == nil {
		t.Fatal("non-canonical group identifier was accepted")
	}
	value = validPersistedRehearsal()
	value.Evidence.Members[1].RunID = value.Evidence.Members[0].RunID
	if _, err := Record(context.Background(), recorder, value); err == nil {
		t.Fatal("cross-member evidence identity reuse was accepted")
	}
}

func validPersistedRehearsal() PersistedRehearsal {
	digest := sha256.Sum256([]byte("persisted-owner-evidence"))
	identities := []uuid.UUID{
		uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		uuid.MustParse("10000000-0000-4000-8000-000000000002"),
		uuid.MustParse("10000000-0000-4000-8000-000000000003"),
		uuid.MustParse("10000000-0000-4000-8000-000000000004"),
		uuid.MustParse("10000000-0000-4000-8000-000000000005"),
		uuid.MustParse("10000000-0000-4000-8000-000000000006"),
	}
	return PersistedRehearsal{
		ReceiptID: uuid.MustParse("10000000-0000-4000-8000-000000000007"),
		Evidence: GroupEvidence{
			TenantID:   uuid.MustParse("10000000-0000-4000-8000-000000000008"),
			Generation: 3, GroupID: CanonicalGroupID,
			QuiesceDrained: true, UsageFlushed: true,
			Members: []MemberEvidence{
				{Domain: "billing", OperationID: identities[0], RunID: identities[1],
					RollbackReceiptID: identities[2], SourceCount: 2, TargetCount: 2,
					AcceptanceSHA256: digest, RollbackSHA256: digest,
					WriterBefore: WriterV1, WriterAfter: WriterV1},
				{Domain: "forms", OperationID: identities[3], RunID: identities[4],
					RollbackReceiptID: identities[5], SourceCount: 4, TargetCount: 4,
					AcceptanceSHA256: digest, RollbackSHA256: digest,
					WriterBefore: WriterV1, WriterAfter: WriterV1},
			},
		},
		RecordedAt: time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC),
	}
}
