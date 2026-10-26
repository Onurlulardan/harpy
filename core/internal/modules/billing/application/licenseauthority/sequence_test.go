package licenseauthority

import (
	"context"
	"errors"
	"testing"
)

func TestSequenceDecisionRejectsRollbackAndConflictingReplay(t *testing.T) {
	fixture := newAuthorityFixture(t)
	snapshot, err := Verify(
		context.Background(), fixture.document(t), fixture.binding, fixture.now,
		fixture.resolver(),
	)
	if err != nil {
		t.Fatal(err)
	}
	head := &AcceptedHead{
		LicenseID: snapshot.LicenseID(), EnvironmentID: snapshot.EnvironmentID(),
		Sequence: snapshot.Sequence(), Digest: snapshot.Digest(),
	}
	decision, err := SequenceDecisionFor(head, snapshot)
	if err != nil || decision != SequenceNoOp {
		t.Fatalf("idempotent replay failed: decision=%q err=%v", decision, err)
	}

	rollback := snapshot
	rollback.sequence--
	if _, err = SequenceDecisionFor(head, rollback); !errors.Is(err, ErrSequenceRollback) {
		t.Fatalf("rollback accepted: %v", err)
	}
	conflict := snapshot
	conflict.digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err = SequenceDecisionFor(head, conflict); !errors.Is(err, ErrSequenceConflict) {
		t.Fatalf("same-sequence conflict accepted: %v", err)
	}
	newer := snapshot
	newer.sequence++
	if decision, err = SequenceDecisionFor(head, newer); err != nil || decision != SequenceAccept {
		t.Fatalf("higher sequence rejected: decision=%q err=%v", decision, err)
	}
}
