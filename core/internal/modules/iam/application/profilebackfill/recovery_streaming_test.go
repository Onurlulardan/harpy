package profilebackfill

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRecoveryAccumulatorPreservesPageWireEvidence(t *testing.T) {
	parent := recoveryFixtureParent()
	accumulator, err := NewRecoveryInventoryAccumulator(
		parent.ID, parent.SourceSchemaID, parent.Target.ID, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	resources := []RecoveryResource{
		recoveryEvidence("14000000-0000-4000-8000-000000000001", parent, "a"),
		recoveryEvidence("14000000-0000-4000-8000-000000000003", parent, "b"),
		recoveryEvidence("14000000-0000-4000-8000-000000000004", parent, "c"),
	}
	if err := accumulator.AppendRemaining(resources[0]); err != nil {
		t.Fatal(err)
	}
	target := SourceResource{
		ID:              uuid.MustParse("14000000-0000-4000-8000-000000000002"),
		SchemaVersionID: parent.Target.ID, Revision: 2, Document: []byte("{}"),
		SHA256: sha256.Sum256([]byte("target-live")),
	}
	if err := accumulator.AppendTarget(target); err != nil {
		t.Fatal(err)
	}
	for _, resource := range resources[1:] {
		if err := accumulator.AppendRemaining(resource); err != nil {
			t.Fatal(err)
		}
	}
	inventory, err := accumulator.Seal()
	if err != nil {
		t.Fatal(err)
	}
	runAt := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	plan, err := PlanRecovery(
		uuid.MustParse("15000000-0000-4000-8000-000000000002"),
		parent, inventory, 2, runAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	wantFirstSource, wantFirstTarget := legacyRecoveryPageDigests(resources[:2])
	wantSecondSource, wantSecondTarget := legacyRecoveryPageDigests(resources[2:])
	pages := plan.Pages()
	if len(pages) != 2 || pages[0].SourceSHA256 != wantFirstSource ||
		pages[0].ExpectedTargetSHA256 != wantFirstTarget ||
		pages[1].SourceSHA256 != wantSecondSource ||
		pages[1].ExpectedTargetSHA256 != wantSecondTarget ||
		pages[0].ThroughResourceID != resources[1].ID ||
		pages[1].AfterResourceID != resources[1].ID ||
		pages[1].ScheduledRunAt != runAt.Add(time.Nanosecond) ||
		inventory.InheritedTargetSHA256() != legacyRecoveryTargetDigest(target) {
		t.Fatalf("streaming pages = %+v", pages)
	}
}

func TestRecoveryAccumulatorRetainsAtMostOneEvidencePage(t *testing.T) {
	parent := recoveryFixtureParent()
	accumulator, err := NewRecoveryInventoryAccumulator(
		parent.ID, parent.SourceSchemaID, parent.Target.ID, MaximumPageSize,
	)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < MaximumPageSize+1; index++ {
		resource := RecoveryResource{
			ID: uuid.MustParse(fmt.Sprintf(
				"14000000-0000-4000-8000-%012x", index+1,
			)),
			SchemaVersionID: parent.SourceSchemaID, Revision: 1,
			SourceSHA256:         sha256.Sum256([]byte(fmt.Sprintf("source:%d", index))),
			ExpectedTargetSHA256: sha256.Sum256([]byte(fmt.Sprintf("target:%d", index))),
		}
		if err := accumulator.AppendRemaining(resource); err != nil {
			t.Fatal(err)
		}
		if len(accumulator.current) > MaximumPageSize {
			t.Fatalf("current evidence length = %d", len(accumulator.current))
		}
	}
	if len(accumulator.pages) != 1 || len(accumulator.current) != 1 {
		t.Fatalf("retained evidence = pages:%d current:%d",
			len(accumulator.pages), len(accumulator.current))
	}
	inventory, err := accumulator.Seal()
	if err != nil || inventory.RemainingSourceCount() != MaximumPageSize+1 ||
		len(accumulator.current) != 0 || accumulator.pages != nil {
		t.Fatalf("finished accumulator = %+v/%v", inventory, err)
	}
}

func recoveryEvidence(id string, parent RecoveryParent, value string) RecoveryResource {
	return RecoveryResource{
		ID: uuid.MustParse(id), SchemaVersionID: parent.SourceSchemaID, Revision: 1,
		SourceSHA256:         sha256.Sum256([]byte("source:" + value)),
		ExpectedTargetSHA256: sha256.Sum256([]byte("target:" + value)),
	}
}

func legacyRecoveryPageDigests(
	resources []RecoveryResource,
) ([sha256.Size]byte, [sha256.Size]byte) {
	sourceDigest, targetDigest := sha256.New(), sha256.New()
	for _, resource := range resources {
		writeRecoveryDigest(sourceDigest, []byte(resource.ID.String()),
			[]byte(resource.SchemaVersionID.String()),
			[]byte(strconv.FormatInt(resource.Revision, 10)), resource.SourceSHA256[:])
		writeRecoveryDigest(targetDigest, []byte(resource.ID.String()),
			resource.ExpectedTargetSHA256[:])
	}
	var source, target [sha256.Size]byte
	copy(source[:], sourceDigest.Sum(nil))
	copy(target[:], targetDigest.Sum(nil))
	return source, target
}

func legacyRecoveryTargetDigest(resource SourceResource) [sha256.Size]byte {
	entry := sha256.New()
	writeRecoveryDigest(entry, []byte(resource.ID.String()),
		[]byte(resource.SchemaVersionID.String()),
		[]byte(strconv.FormatInt(resource.Revision, 10)), resource.SHA256[:])
	entrySHA := entry.Sum(nil)
	initial := sha256.Sum256(nil)
	payload := append(append([]byte{}, initial[:]...), entrySHA...)
	return sha256.Sum256(payload)
}
