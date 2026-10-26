package billingv1adoptionapp

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRehearseRollbackRequiresZeroDifference(t *testing.T) {
	plan := emptyEncodedPlan()
	receiptID := uuid.New()
	at := time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC)
	digest := strings.Repeat("ab", 32)
	store := &fakeScopeStore{rehearsal: RollbackRehearsal{ReceiptID: receiptID,
		ReverseSHA256: digest, SourceSHA256: digest, RehearsedAt: at}}
	if _, err := RehearseRollback(context.Background(), store, scopeOperation(plan),
		receiptID, digest, at); err != nil {
		t.Fatal(err)
	}
	store.rehearsal.ReverseSHA256 = strings.Repeat("cd", 32)
	if _, err := RehearseRollback(context.Background(), store, scopeOperation(plan),
		receiptID, digest, at); err == nil {
		t.Fatal("reverse difference was accepted")
	}
}
