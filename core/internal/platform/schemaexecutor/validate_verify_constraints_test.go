package schemaexecutor

import (
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestVerifyConstraintsRequireExactMappingCoverage(t *testing.T) {
	mapping := schemav1.IdentifierMapping{Constraints: []schemav1.Identifier{{
		LogicalID:    "00000000-0000-4000-8000-000000000008",
		PhysicalName: "cst_example",
	}}}
	arguments := schemav1.VerifyArguments{
		Checks: []schemav1.VerifyCheck{schemav1.VerifyConstraints},
		Fields: []schemav1.VerifyFieldContract{}, Relations: []schemav1.RelationArguments{},
		Indexes:     []schemav1.VerifyIndexContract{},
		Constraints: []schemav1.VerifyConstraintContract{},
	}
	if err := validateVerifyArguments(mapping, arguments); err == nil {
		t.Fatal("expected missing mapped constraint rejection")
	}
}
