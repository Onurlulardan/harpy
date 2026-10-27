package grouprehearsal

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillingOnlyGroupRehearsalIsValid(t *testing.T) {
	value := billingRehearsalEvidence(MandatoryGroupID)
	if _, err := EvaluateGroup(value); err != nil {
		t.Fatal(err)
	}
	persisted := PersistedRehearsal{
		ReceiptID:  uuid.MustParse("71000000-0000-4000-8000-000000000021"),
		Evidence:   billingRehearsalEvidence(CanonicalGroupID),
		RecordedAt: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
	}
	if err := validatePersisted(persisted); err != nil {
		t.Fatal(err)
	}
}

func TestFormsCompanionRehearsalIsRejected(t *testing.T) {
	value := billingRehearsalEvidence(MandatoryGroupID)
	forms := value.Members[0]
	forms.Domain = "forms"
	forms.OperationID = uuid.MustParse("71000000-0000-4000-8000-000000000031")
	forms.RunID = uuid.MustParse("71000000-0000-4000-8000-000000000032")
	forms.RollbackReceiptID = uuid.MustParse("71000000-0000-4000-8000-000000000033")
	value.Members = append(value.Members, forms)
	if _, err := EvaluateGroup(value); err == nil {
		t.Fatal("forms companion rehearsal was accepted")
	}
	persisted := PersistedRehearsal{
		ReceiptID:  uuid.MustParse("71000000-0000-4000-8000-000000000021"),
		Evidence:   value,
		RecordedAt: time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
	}
	persisted.Evidence.GroupID = CanonicalGroupID
	if err := validatePersisted(persisted); err == nil {
		t.Fatal("persisted forms companion rehearsal was accepted")
	}
}

func billingRehearsalEvidence(groupID string) GroupEvidence {
	return GroupEvidence{
		TenantID:       uuid.MustParse("71000000-0000-4000-8000-000000000001"),
		Generation:     1,
		GroupID:        groupID,
		QuiesceDrained: true,
		UsageFlushed:   true,
		Members: []MemberEvidence{{
			Domain:            "billing",
			OperationID:       uuid.MustParse("71000000-0000-4000-8000-000000000011"),
			RunID:             uuid.MustParse("71000000-0000-4000-8000-000000000012"),
			RollbackReceiptID: uuid.MustParse("71000000-0000-4000-8000-000000000013"),
			SourceCount:       3,
			TargetCount:       3,
			AcceptanceSHA256:  [32]byte{1},
			RollbackSHA256:    [32]byte{2},
			WriterBefore:      WriterV1,
			WriterAfter:       WriterV1,
		}},
	}
}
