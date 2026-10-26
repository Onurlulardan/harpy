package schemav1

import (
	"strings"
	"testing"
)

func TestPreparePlanPinsCanonicalStepWrapper(t *testing.T) {
	plan := preparationFixture()
	prepared, err := PreparePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.SHA256() != "1e5b43107bb0a368982f92dd9bdaa100190ad327d15b04c35fe62e3211c06dd5" {
		t.Fatalf("plan digest = %s", prepared.SHA256())
	}
	wantPlan := `{"allowed_operations":["create_form_table"],` +
		`"compiler":{"postgresql_major":18,"version":"1.0.0"},` +
		`"contract_version":"forms-schema-plan.v1","owner":"forms",` +
		`"resource":{"id":"018f47f4-7b1b-7cc0-8000-000000000702",` +
		`"kind":"form_storage","revision":1},"steps":[` +
		`{"arguments":{},"executor_kind":"migrator",` +
		`"identifier_mapping":{"constraints":[],"fields":[],` +
		`"form":{"logical_id":"018f47f4-7b1b-7cc0-8000-000000000702",` +
		`"physical_name":"frm_aghup5d3dn6mbaaaaaaaaaahai"},` +
		`"indexes":[],"related_forms":[],"relations":[]},` +
		`"operation_id":"create_form_table","ordinal":0,` +
		`"transaction_class":"transactional"}],` +
		`"tenant_id":"018f47f4-7b1b-7cc0-8000-000000000701"}`
	if string(prepared.CanonicalJSON()) != wantPlan {
		t.Fatalf("canonical plan = %s", prepared.CanonicalJSON())
	}
	steps := prepared.Steps()
	if len(steps) != 1 || steps[0].Ordinal() != 0 {
		t.Fatalf("prepared steps = %#v", steps)
	}
	wantPreimage := `{"plan_sha256":"` + prepared.SHA256() +
		`","step":` + wantPlanStep() + `}`
	if string(steps[0].CanonicalPreimage()) != wantPreimage {
		t.Fatalf("step preimage = %s", steps[0].CanonicalPreimage())
	}
	if steps[0].SHA256() != "9b89f2d5172787daf98a764223fa2e2324a21211317d84cad8fe113c91b07cf7" {
		t.Fatalf("step digest = %s", steps[0].SHA256())
	}
	copyBytes := steps[0].CanonicalPreimage()
	copyBytes[0] = '['
	if steps[0].CanonicalPreimage()[0] != '{' {
		t.Fatal("prepared step aliases caller bytes")
	}
}

func wantPlanStep() string {
	return `{"arguments":{},"executor_kind":"migrator",` +
		`"identifier_mapping":{"constraints":[],"fields":[],` +
		`"form":{"logical_id":"018f47f4-7b1b-7cc0-8000-000000000702",` +
		`"physical_name":"frm_aghup5d3dn6mbaaaaaaaaaahai"},` +
		`"indexes":[],"related_forms":[],"relations":[]},` +
		`"operation_id":"create_form_table","ordinal":0,` +
		`"transaction_class":"transactional"}`
}

func TestPreparePlanRejectsNonGaplessOrOpenDigest(t *testing.T) {
	plan := preparationFixture()
	plan.Steps[0].Ordinal = 1
	if _, err := PreparePlan(plan); err == nil {
		t.Fatal("non-gapless plan accepted")
	}
	if _, err := PrepareStep(strings.Repeat("z", 64), Step{}); err == nil {
		t.Fatal("open digest accepted")
	}
}

func preparationFixture() Plan {
	const id = "018f47f4-7b1b-7cc0-8000-000000000702"
	name, _ := DeriveIdentifier(IdentifierForm, id)
	mapping := IdentifierMapping{
		Form:         Identifier{LogicalID: id, PhysicalName: name},
		RelatedForms: []Identifier{}, Fields: []FieldIdentifier{},
		Relations: []Identifier{}, Indexes: []Identifier{}, Constraints: []Identifier{},
	}
	return Plan{
		ContractVersion: PlanContractVersion,
		TenantID:        "018f47f4-7b1b-7cc0-8000-000000000701", Owner: "forms",
		Resource: Resource{Kind: ResourceFormStorage, ID: id, Revision: 1},
		Compiler: Compiler{
			Version: InitialCompilerVersion, PostgreSQLMajor: RequiredPostgreSQLMajor,
		},
		AllowedOperations: []OperationID{CreateFormTable},
		Steps: []Step{{
			Ordinal: 0, OperationID: CreateFormTable,
			TransactionClass: Transactional, ExecutorKind: MigratorExecutor,
			IdentifierMapping: mapping, Arguments: OperationArguments{},
		}},
	}
}
