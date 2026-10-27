package billingv1adoptionapp

import (
	"context"
	"testing"
)

func TestVerifyScopeRejectsAnyDifference(t *testing.T) {
	plan := emptyEncodedPlan()
	digest := plan.SourceSHA256
	store := &fakeScopeStore{verified: Reconciliation{SourceCount: 1,
		TargetCount: 1, SourceSHA256: digest, TargetSHA256: digest}}
	if _, err := VerifyScope(context.Background(), store, scopeOperation(plan)); err != nil {
		t.Fatal(err)
	}
	store.verified.UsageDifference = 1
	if _, err := VerifyScope(context.Background(), store, scopeOperation(plan)); err == nil {
		t.Fatal("usage difference was accepted")
	}
}
