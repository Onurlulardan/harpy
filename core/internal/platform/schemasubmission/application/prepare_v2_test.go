package application

import (
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/google/uuid"
)

func TestPrepareDispatchesExactV2Pair(t *testing.T) {
	plan := fixturePlanV2()
	prepared, err := prepare(SubmitCommand{
		ProposedJobID: uuid.MustParse(fixtureJobID), PlanV2: &plan,
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.ContractVersion() != schemav2.PlanContractVersion ||
		prepared.CompilerVersion() != schemav2.CompilerVersion ||
		len(prepared.CanonicalSteps()) != 1 {
		t.Fatalf("prepared V2 identity differs: %+v", prepared)
	}
}

func TestPrepareRejectsMixedV1AndV2Plan(t *testing.T) {
	v2Plan := fixturePlanV2()
	if _, err := prepare(SubmitCommand{
		ProposedJobID: uuid.MustParse(fixtureJobID),
		Plan:          fixturePlan(), PlanV2: &v2Plan,
	}); err == nil {
		t.Fatal("mixed plan contracts accepted")
	}
}

func fixturePlanV2() schemav2.Plan {
	v1 := fixturePlan()
	return schemav2.Plan{
		ContractVersion: schemav2.PlanContractVersion,
		TenantID:        v1.TenantID, Owner: v1.Owner,
		Resource: schemav2.Resource{
			Kind: schemav2.ResourceFormStorage,
			ID:   v1.Resource.ID, Revision: v1.Resource.Revision,
		},
		Compiler: schemav2.Compiler{
			Version:         schemav2.CompilerVersion,
			PostgreSQLMajor: schemav2.RequiredPostgreSQLMajor,
		},
		AllowedOperations: []schemav2.OperationID{schemav2.CreateFormTable},
		Steps: []schemav2.Step{{
			Ordinal: 0, OperationID: schemav2.CreateFormTable,
			TransactionClass: schemav2.Transactional,
			ExecutorKind:     schemav2.MigratorExecutor,
			IdentifierMapping: schemav2.IdentifierMapping{
				Form: schemav2.Identifier{
					LogicalID:    v1.Resource.ID,
					PhysicalName: fixturePhysicalName("frm_", v1.Resource.ID),
				},
				RelatedForms: []schemav2.Identifier{}, Fields: []schemav2.FieldIdentifier{},
				Relations: []schemav2.Identifier{}, Indexes: []schemav2.Identifier{},
				Constraints: []schemav2.Identifier{},
			},
			Arguments: schemav2.OperationArguments{},
		}},
	}
}
