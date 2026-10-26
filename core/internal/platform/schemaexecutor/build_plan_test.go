package schemaexecutor

import (
	"bytes"
	"testing"
	"time"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestBuildPlanIsDeterministic(t *testing.T) {
	plan := testCreateFormPlan(t)
	first, err := BuildPlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildPlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	if first.SHA256() != second.SHA256() ||
		!bytes.Equal(first.CanonicalJSON(), second.CanonicalJSON()) ||
		first.Steps()[0].SHA256() != second.Steps()[0].SHA256() {
		t.Fatal("equivalent typed plans produced different identities")
	}
}

func TestBuildPlanMatchesSharedPreparationContract(t *testing.T) {
	plan := testCreateFormPlan(t)
	shared, err := schemav1.PreparePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := BuildPlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	sharedStep, platformStep := shared.Steps()[0], prepared.Steps()[0]
	if prepared.SHA256() != shared.SHA256() ||
		!bytes.Equal(prepared.CanonicalJSON(), shared.CanonicalJSON()) ||
		platformStep.SHA256() != sharedStep.SHA256() ||
		!bytes.Equal(platformStep.CanonicalPreimage(), sharedStep.CanonicalPreimage()) {
		t.Fatal("platform preparation drifted from shared schema contract")
	}
}

func TestValidateClaimRejectsPlanTamper(t *testing.T) {
	record := testClaimRecord(t, testClock())
	record.Job.CanonicalPlan = append(record.Job.CanonicalPlan, ' ')
	claim, err := acceptClaim(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateClaim(claim); err != nil {
		t.Fatalf("whitespace must preserve canonical identity: %v", err)
	}
	record.Job.CanonicalPlan[1] = 'x'
	claim, err = acceptClaim(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateClaim(claim); err == nil {
		t.Fatal("tampered plan was accepted")
	}
}

func testClock() time.Time {
	return time.Date(2026, time.July, 14, 10, 0, 0, 0, time.UTC)
}
