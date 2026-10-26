package schemav2

import (
	"encoding/json"
	"testing"
)

const (
	testTenantID = "018f47f4-7b1b-7cc0-8000-000000000701"
	testFormID   = "018f47f4-7b1b-7cc0-8000-000000000702"
	testFieldID  = "018f47f4-7b1b-7cc0-8000-000000000703"
)

func preparationFixture(t *testing.T) Plan {
	t.Helper()
	formName, err := DeriveIdentifier(IdentifierForm, testFormID)
	if err != nil {
		t.Fatal(err)
	}
	foldedName, err := DeriveFieldIdentifier(testFieldID, 1, VariantFolded)
	if err != nil {
		t.Fatal(err)
	}
	empty := IdentifierMapping{
		Form:         Identifier{LogicalID: testFormID, PhysicalName: formName},
		RelatedForms: []Identifier{}, Fields: []FieldIdentifier{},
		Relations: []Identifier{}, Indexes: []Identifier{}, Constraints: []Identifier{},
	}
	folded := empty
	folded.Fields = []FieldIdentifier{{
		LogicalID: testFieldID, Generation: 1,
		Variant: VariantFolded, PhysicalName: foldedName,
	}}
	return Plan{
		ContractVersion: PlanContractVersion, TenantID: testTenantID, Owner: "forms",
		Resource:          Resource{Kind: ResourceFormStorage, ID: testFormID, Revision: 7},
		Compiler:          Compiler{Version: CompilerVersion, PostgreSQLMajor: RequiredPostgreSQLMajor},
		AllowedOperations: []OperationID{AddFieldColumn, CreateFormTable},
		Steps: []Step{
			{Ordinal: 0, OperationID: CreateFormTable, TransactionClass: Transactional,
				ExecutorKind: MigratorExecutor, IdentifierMapping: empty,
				Arguments: OperationArguments{}},
			{Ordinal: 1, OperationID: AddFieldColumn, TransactionClass: Transactional,
				ExecutorKind: MigratorExecutor, IdentifierMapping: folded,
				Arguments: arguments(t, ColumnArguments{
					FieldID: testFieldID, Generation: 1, Variant: VariantFolded,
					StorageType: StorageText, Nullable: true,
				})},
		},
	}
}

func arguments(t *testing.T, value any) OperationArguments {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	result := OperationArguments{}
	if err := json.Unmarshal(document, &result); err != nil {
		t.Fatal(err)
	}
	return result
}
