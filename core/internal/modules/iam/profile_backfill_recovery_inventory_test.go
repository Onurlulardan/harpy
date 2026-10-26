package iam

import (
	"errors"
	"testing"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func TestProfileRecoveryInventoryRejectsUnknownSchema(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 64)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 64)
	unknownID := uuid.MustParse("13000000-0000-4000-8000-000000000003")
	_, err := classifyProfileRecoveryFixture(
		t, profileRecoveryParent(t, source, target), source, target,
		map[uuid.UUID]profileapp.Schema{},
		[]profilebackfillapp.SourceResource{profileRecoveryResource(
			"14000000-0000-4000-8000-000000000001", unknownID, []byte("{}"),
		)},
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("unknown schema error = %v", err)
	}
}

func TestProfileRecoveryInventoryRejectsInvalidExactTarget(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 64)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 3)
	document := []byte(`{"12000000-0000-4000-8000-000000000001":"Alice"}`)
	_, err := classifyProfileRecoveryFixture(
		t, profileRecoveryParent(t, source, target), source, target, nil,
		[]profilebackfillapp.SourceResource{profileRecoveryResource(
			"14000000-0000-4000-8000-000000000001", target.ID(), document,
		)},
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("invalid target error = %v", err)
	}
}

func TestProfileRecoveryInventoryAllowsZeroRemaining(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000001", 1, fieldID, false, 64)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 64)
	document := []byte(`{"12000000-0000-4000-8000-000000000001":"Alice"}`)
	parent := profileRecoveryParent(t, source, target)
	inventory, err := classifyProfileRecoveryFixture(
		t, parent, source, target, nil,
		[]profilebackfillapp.SourceResource{profileRecoveryResource(
			"14000000-0000-4000-8000-000000000001", target.ID(), document,
		)},
	)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := profilebackfillapp.PlanRecovery(
		uuid.MustParse("15000000-0000-4000-8000-000000000002"), parent, inventory,
		profilebackfillapp.MaximumPageSize, time.Unix(1, 0).UTC(),
	)
	if err != nil || inventory.InheritedTargetCount() != 1 ||
		inventory.RemainingSourceCount() != 0 || len(plan.Pages()) != 0 {
		t.Fatalf("zero remaining = %d/%d/%d/%v", inventory.InheritedTargetCount(),
			inventory.RemainingSourceCount(), len(plan.Pages()), err)
	}
}
