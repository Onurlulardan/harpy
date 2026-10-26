package schemaexecutor

import (
	"bytes"
	"encoding/json"
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

const testFieldIDV2 = "00000000-0000-4000-8000-000000000006"
const testIndexIDV2 = "00000000-0000-4000-8000-000000000007"

func TestBuildPlanV2MatchesSharedPreparationContract(t *testing.T) {
	plan := testTrigramPlanV2(t)
	shared, err := schemav2.PreparePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := BuildPlanV2(plan)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.SHA256() != shared.SHA256() ||
		!bytes.Equal(prepared.CanonicalJSON(), shared.CanonicalJSON()) {
		t.Fatal("V2 executor preparation drifted from contract")
	}
	for index, step := range shared.Steps() {
		actual := prepared.Steps()[index]
		if actual.SHA256() != step.SHA256() ||
			!bytes.Equal(actual.CanonicalPreimage(), step.CanonicalPreimage()) {
			t.Fatalf("V2 step %d preparation drifted", index)
		}
	}
}

func TestBuildPlanV2RejectsVariantAndOperatorClassInjection(t *testing.T) {
	for name, mutate := range map[string]func(*schemav2.Plan){
		"operator class": func(plan *schemav2.Plan) {
			plan.Steps[2].Arguments["operator_class"] = json.RawMessage(`"gin_trgm_ops;DROP"`)
		},
		"wrong variant": func(plan *schemav2.Plan) {
			plan.Steps[2].Arguments["fields"] = json.RawMessage(
				`[{"field_id":"` + testFieldIDV2 + `","generation":1,` +
					`"variant":"value","storage_type":"text"}]`,
			)
		},
		"accent without folded": func(plan *schemav2.Plan) {
			field := &plan.Steps[1].IdentifierMapping.Fields[1]
			field.Variant = schemav2.VariantAccent
			field.PhysicalName, _ = schemav2.DeriveFieldIdentifier(
				field.LogicalID, field.Generation, field.Variant,
			)
		},
	} {
		t.Run(name, func(t *testing.T) {
			plan := testTrigramPlanV2(t)
			mutate(&plan)
			if _, err := BuildPlanV2(plan); err == nil {
				t.Fatal("invalid V2 plan accepted")
			}
		})
	}
}

func testTrigramPlanV2(t *testing.T) schemav2.Plan {
	t.Helper()
	formName, _ := schemav2.DeriveIdentifier(schemav2.IdentifierForm, testFormID)
	valueName, _ := schemav2.DeriveFieldIdentifier(
		testFieldIDV2, 1, schemav2.VariantValue,
	)
	foldedName, _ := schemav2.DeriveFieldIdentifier(
		testFieldIDV2, 1, schemav2.VariantFolded,
	)
	indexName, _ := schemav2.DeriveIndexIdentifier(testIndexIDV2, 1)
	base := schemav2.IdentifierMapping{
		Form:         schemav2.Identifier{LogicalID: testFormID, PhysicalName: formName},
		RelatedForms: []schemav2.Identifier{}, Fields: []schemav2.FieldIdentifier{},
		Relations: []schemav2.Identifier{}, Indexes: []schemav2.Identifier{},
		Constraints: []schemav2.Identifier{},
	}
	fields := base
	fields.Fields = []schemav2.FieldIdentifier{
		{LogicalID: testFieldIDV2, Generation: 1,
			Variant: schemav2.VariantValue, PhysicalName: valueName},
		{LogicalID: testFieldIDV2, Generation: 1,
			Variant: schemav2.VariantFolded, PhysicalName: foldedName},
	}
	index := fields
	index.Indexes = []schemav2.Identifier{{
		LogicalID: testIndexIDV2, Generation: 1, PhysicalName: indexName,
	}}
	return schemav2.Plan{
		ContractVersion: schemav2.PlanContractVersion,
		TenantID:        testTenantID, Owner: "forms",
		Resource: schemav2.Resource{
			Kind: schemav2.ResourceFormStorage, ID: testFormID, Revision: 1,
		},
		Compiler: schemav2.Compiler{
			Version:         schemav2.CompilerVersion,
			PostgreSQLMajor: schemav2.RequiredPostgreSQLMajor,
		},
		AllowedOperations: []schemav2.OperationID{
			schemav2.AddFieldColumn, schemav2.CreateFormTable, schemav2.CreateIndex,
		},
		Steps: []schemav2.Step{
			v2Step(t, 0, schemav2.CreateFormTable, base, schemav2.EmptyArguments{}),
			v2Step(t, 1, schemav2.AddFieldColumn, fields, schemav2.ColumnArguments{
				FieldID: testFieldIDV2, Generation: 1, Variant: schemav2.VariantFolded,
				StorageType: schemav2.StorageText, Nullable: true,
			}),
			v2Step(t, 2, schemav2.CreateIndex, index, schemav2.IndexArguments{
				IndexID: testIndexIDV2, Generation: 1,
				Fields: []schemav2.IndexField{{
					FieldID: testFieldIDV2, Generation: 1,
					Variant: schemav2.VariantFolded, StorageType: schemav2.StorageText,
				}},
				Method: schemav2.IndexGIN, OperatorClass: schemav2.OperatorClassGINTrigram,
				ActiveRecordsOnly: true,
			}),
		},
	}
}

func v2Step(
	t *testing.T,
	ordinal int,
	operation schemav2.OperationID,
	mapping schemav2.IdentifierMapping,
	value any,
) schemav2.Step {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	arguments := schemav2.OperationArguments{}
	if string(document) != "{}" {
		if err := json.Unmarshal(document, &arguments); err != nil {
			t.Fatal(err)
		}
	}
	return schemav2.Step{
		Ordinal: ordinal, OperationID: operation, TransactionClass: schemav2.Transactional,
		ExecutorKind:      schemav2.MigratorExecutor,
		IdentifierMapping: mapping, Arguments: arguments,
	}
}
