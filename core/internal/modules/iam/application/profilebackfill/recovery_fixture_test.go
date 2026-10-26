package profilebackfill

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func recoveryFixtureParent() RecoveryParent {
	sourceID := uuid.MustParse("13000000-0000-4000-8000-000000000001")
	targetID := uuid.MustParse("13000000-0000-4000-8000-000000000002")
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	definition := []byte("{}")
	return RecoveryParent{
		ID:    uuid.MustParse("15000000-0000-4000-8000-000000000001"),
		State: RecoveryParentFenced, PlanSHA256: sha256.Sum256([]byte("parent")),
		SourceSchemaID: sourceID,
		Target: TargetSchema{
			ID: targetID, ResourceKind: "user", Version: 2,
			DefinitionJSON: definition, DefinitionSHA256: sha256.Sum256(definition),
			Fields: []TargetField{{
				ID: fieldID, Active: true,
				SemanticSHA256: sha256.Sum256([]byte("field")),
			}},
		},
		ConverterVersion: ConverterTypedMap,
		Mappings: []Mapping{{
			SourceFieldID: fieldID, TargetFieldID: fieldID,
			ConverterKind: MappingCopy,
		}},
	}
}

func recoveryFixtureInventory(
	t *testing.T,
	parent RecoveryParent,
	remaining int,
) RecoveryInventory {
	t.Helper()
	accumulator, err := NewRecoveryInventoryAccumulator(
		parent.ID, parent.SourceSchemaID, parent.Target.ID, MaximumPageSize,
	)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < remaining; index++ {
		resource := RecoveryResource{
			ID: uuid.MustParse(fmt.Sprintf(
				"14000000-0000-4000-8000-%012x", index+1,
			)),
			SchemaVersionID: parent.SourceSchemaID, Revision: int64(index + 1),
			SourceSHA256:         sha256.Sum256([]byte("source")),
			ExpectedTargetSHA256: sha256.Sum256([]byte("target")),
		}
		if err := accumulator.AppendRemaining(resource); err != nil {
			t.Fatal(err)
		}
	}
	inventory, err := accumulator.Seal()
	if err != nil {
		t.Fatal(err)
	}
	return inventory
}

func recoveryFixturePlan(
	t *testing.T,
	parent RecoveryParent,
	remaining int,
) RecoveryPlan {
	t.Helper()
	plan, err := PlanRecovery(
		uuid.MustParse("15000000-0000-4000-8000-000000000002"),
		parent, recoveryFixtureInventory(t, parent, remaining), MaximumPageSize,
		time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func recoveryFixtureEvidence() Evidence {
	id := uuid.MustParse("16000000-0000-4000-8000-000000000001")
	return Evidence{
		OperationID: id, RequestSHA256: sha256.Sum256([]byte("request")),
		ActorID: id, ManagePermissionID: id,
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
}
