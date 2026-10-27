package billingv1adoptionapp

import (
	"testing"
)

func TestCompileBlocksUnprovenCouponAndOpeningEvidence(t *testing.T) {
	snapshot, mapping := validAdoptionFixture()
	snapshot.Global.Coupons[0].ApplicablePlanIDs = nil
	mapping.Tenants[0].UsageFlushEvidenceSHA256 = nil
	snapshot.Tenants[0].Gauges = nil
	bundle, err := Compile(snapshot, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Global.Eligible || bundle.Tenants[0].Eligible {
		t.Fatal("unproven Billing V1 evidence was eligible")
	}
	assertIssueCode(t, bundle.Global.Issues, "global_applicability_unproven")
	assertIssueCode(t, bundle.Tenants[0].Issues, "opening_evidence_invalid")
	assertIssueCode(t, bundle.Tenants[0].Issues, "owner_evidence_missing")
}

func TestCompileBlocksPastDueWithoutGraceEvidence(t *testing.T) {
	snapshot, mapping := validAdoptionFixture()
	snapshot.Tenants[0].Subscriptions[0].State = "past_due"
	bundle, err := Compile(snapshot, mapping)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Tenants[0].Eligible {
		t.Fatal("past_due without grace evidence was eligible")
	}
	assertIssueCode(t, bundle.Tenants[0].Issues, "past_due_grace_unproven")
}

func TestCompileBlocksCouponUseCountDrift(t *testing.T) {
	snapshot, mapping := validAdoptionFixture()
	snapshot.Global.Coupons[0].UsedCount = 2
	bundle, err := Compile(snapshot, mapping)
	if err != nil {
		t.Fatal(err)
	}
	assertIssueCode(t, bundle.Global.Issues, "global_use_count_mismatch")
}

func assertIssueCode(t *testing.T, issues []Issue, code string) {
	t.Helper()
	for _, value := range issues {
		if value.Code == code {
			return
		}
	}
	t.Fatalf("issue code %q missing from %+v", code, issues)
}
