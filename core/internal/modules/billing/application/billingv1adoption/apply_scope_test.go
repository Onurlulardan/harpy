package billingv1adoptionapp

import (
	"context"
	"testing"
)

func TestApplyScopeReturnsPersistedCount(t *testing.T) {
	plan := emptyEncodedPlan()
	count, err := ApplyScope(context.Background(), &fakeScopeStore{applied: 7},
		scopeOperation(plan))
	if err != nil || count != 7 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}
