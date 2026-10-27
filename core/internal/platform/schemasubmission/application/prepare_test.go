package application

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestPreparePinsCanonicalPlanAndStepPreimages(t *testing.T) {
	prepared, err := prepare(SubmitCommand{
		ProposedJobID: uuid.MustParse(fixtureJobID), Plan: fixturePlan(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.TenantID().String() != fixtureTenantID ||
		prepared.ResourceID().String() != fixtureResourceID ||
		prepared.PlanSHA256() == "" || len(prepared.CanonicalSteps()) != 1 {
		t.Fatalf("prepared identity differs: %+v", prepared)
	}
	planBytes := prepared.CanonicalPlan()
	planBytes[0] = 'x'
	if bytes.Equal(planBytes, prepared.CanonicalPlan()) {
		t.Fatal("canonical plan accessor leaked mutable storage")
	}
	steps := prepared.CanonicalSteps()
	steps[0][0] = 'x'
	if bytes.Equal(steps[0], prepared.CanonicalSteps()[0]) {
		t.Fatal("canonical step accessor leaked mutable storage")
	}
}

func TestPrepareRejectsNonUUIDv7AndInvalidPlan(t *testing.T) {
	plan := fixturePlan()
	plan.TenantID = "018f4f9e-7c21-4a3b-8c4d-123456789abc"
	if _, err := prepare(SubmitCommand{
		ProposedJobID: uuid.MustParse(fixtureJobID), Plan: plan,
	}); err == nil {
		t.Fatal("non-UUIDv7 tenant accepted")
	}
	plan = fixturePlan()
	plan.Steps[0].ExecutorKind = "forms_runtime"
	if _, err := prepare(SubmitCommand{
		ProposedJobID: uuid.MustParse(fixtureJobID), Plan: plan,
	}); err == nil {
		t.Fatal("invalid executor contract accepted")
	}
}
