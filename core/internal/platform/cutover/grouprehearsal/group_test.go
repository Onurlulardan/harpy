package grouprehearsal

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func TestBillingFormsGroupRequiresAtomicZeroDifferenceRollback(t *testing.T) {
	value := validGroupEvidence()
	accepted, err := EvaluateGroup(value)
	if err != nil {
		t.Fatal(err)
	}
	if digest, _ := accepted.SHA256(); digest == [32]byte{} {
		t.Fatal("empty group digest")
	}
	value.Members[1].WriterAfter = "v2"
	if _, err := EvaluateGroup(value); err == nil {
		t.Fatal("partial Forms switch accepted")
	}
	value = validGroupEvidence()
	value.ProductionDualWrite = true
	if _, err := EvaluateGroup(value); err == nil {
		t.Fatal("production dual write accepted")
	}
}

func validGroupEvidence() GroupEvidence {
	digest := sha256.Sum256([]byte("evidence"))
	return GroupEvidence{TenantID: uuid.MustParse("01910000-0000-7000-8000-000000000001"),
		Generation: 3, GroupID: MandatoryGroupID, QuiesceDrained: true, UsageFlushed: true,
		Members: []MemberEvidence{
			{Domain: "billing", SourceCount: 2, TargetCount: 2, AcceptanceSHA256: digest,
				RollbackSHA256: digest, WriterBefore: WriterV1, WriterAfter: WriterV1},
			{Domain: "forms", SourceCount: 4, TargetCount: 4, AcceptanceSHA256: digest,
				RollbackSHA256: digest, WriterBefore: WriterV1, WriterAfter: WriterV1},
		}}
}
