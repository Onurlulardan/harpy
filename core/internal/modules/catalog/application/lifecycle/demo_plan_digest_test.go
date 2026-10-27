package lifecycle

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestSealDemoRemovalPlanIsDeterministicAndBounded(t *testing.T) {
	installationID := uuid.MustParse("018f4000-0000-7000-8000-000000000001")
	componentA := uuid.MustParse("018f4000-0000-7000-8000-000000000002")
	componentB := uuid.MustParse("018f4000-0000-7000-8000-000000000003")
	runA := uuid.MustParse("018f4000-0000-7000-8000-000000000004")
	runB := uuid.MustParse("018f4000-0000-7000-8000-000000000005")
	recordA := uuid.MustParse("76000000-0000-4000-8000-000000000001")
	recordB := uuid.MustParse("76000000-0000-4000-8000-000000000002")
	left, err := sealDemoRemovalPlan(installationID, []DemoRemovalComponent{
		{ComponentID: componentB, SourceRunID: runB, RecordIDs: []uuid.UUID{recordB}},
		{ComponentID: componentA, SourceRunID: runA, RecordIDs: []uuid.UUID{recordB, recordA}},
	})
	if err != nil {
		t.Fatal(err)
	}
	right, err := sealDemoRemovalPlan(installationID, []DemoRemovalComponent{
		{ComponentID: componentA, SourceRunID: runA, RecordIDs: []uuid.UUID{recordA, recordB}},
		{ComponentID: componentB, SourceRunID: runB, RecordIDs: []uuid.UUID{recordB}},
	})
	if err != nil || left.PlanSHA256 != right.PlanSHA256 ||
		left.Components[0].ComponentID != componentA ||
		left.Components[0].RecordIDs[0] != recordA {
		t.Fatalf("deterministic plans differ: %#v %#v err=%v", left, right, err)
	}
	_, err = sealDemoRemovalPlan(installationID, []DemoRemovalComponent{{
		ComponentID: componentA, SourceRunID: runA,
		RecordIDs: []uuid.UUID{recordA, recordA},
	}})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("duplicate record error=%v", err)
	}
}
