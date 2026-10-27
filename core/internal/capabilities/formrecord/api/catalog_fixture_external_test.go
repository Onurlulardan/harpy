package recordkernel

import (
	"crypto/sha256"
	"encoding/json"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const (
	tenantID      = "71000000-0000-4000-8000-000000001001"
	publicationID = "018f47f4-7b1b-7cc0-8000-000000001002"
	formID        = "018f47f4-7b1b-7cc0-8000-000000001003"
	versionID     = "018f47f4-7b1b-7cc0-8000-000000001004"
	fieldID       = "018f47f4-7b1b-7cc0-8000-000000001005"
	countFieldID  = "018f47f4-7b1b-7cc0-8000-000000001006"
	calcFieldID   = "018f47f4-7b1b-7cc0-8000-000000001007"
	ruleID        = "018f47f4-7b1b-7cc0-8000-000000001008"
)

func fixtureDefinition(t *testing.T, contract string) []byte {
	t.Helper()
	add, countReference := "add", countFieldID
	schema := StorageSchema{
		FieldRegistryContract: RegistryContractVersion,
		Fields: []FieldDefinition{
			{ID: fieldID, Name: "title", TypeID: TypeText,
				Role: RoleStored, Nullable: false,
				StorageFeatures: []StorageFeature{FeatureBaseColumn}},
			{ID: countFieldID, Name: "count", TypeID: TypeID(TypeInteger),
				Role: RoleStored, Nullable: false,
				StorageFeatures: []StorageFeature{FeatureBaseColumn}},
			{ID: calcFieldID, Name: "total", TypeID: TypeID(TypeInteger),
				Role: RoleCalculation, Nullable: true,
				StorageFeatures: []StorageFeature{FeatureBaseColumn}},
		}, Relations: []RelationDefinition{}, Indexes: []IndexDefinition{},
		Constraints: []ConstraintDefinition{},
		Validations: []ValidationDefinition{},
		Calculations: []CalculationDefinition{{FieldID: calcFieldID,
			Expression: Expression{Kind: "binary", Operator: &add,
				Arguments: []Expression{
					{Kind: "field", FieldID: &countReference},
					{Kind: "literal", Literal: json.RawMessage(`1`)},
				}}}},
	}
	document := map[string]any{
		"contract_version": contract, "form_version_id": versionID,
		"schema_contract": StorageSchemaContractVersion,
		"schema":          schema, "presentation_contract": PresentationContractVersion,
		"presentation": map[string]any{"layout": "default"},
	}
	if contract == DefinitionV2Contract {
		document["record_contract"] = RecordRulesContract
		document["record"] = RecordRules{
			Fields: []RecordFieldRule{
				{FieldID: fieldID, WritePolicy: WriteActorWorkflow},
				{FieldID: countFieldID, WritePolicy: WriteActorWorkflow},
				{FieldID: calcFieldID, WritePolicy: WriteEngine},
			}, WorkflowActions: []string{"create", "update"},
			Defaults: []DefaultRule{{FieldID: countFieldID,
				Expression: typedLiteral("integer", `"1"`)}},
			CrossFieldValidations: []CrossFieldValidation{{
				ID: ruleID, Code: "TOTAL_POSITIVE", Predicate: typedGTTotal(),
			}},
		}
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, 4<<20)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

func fixtureCompileInput(t *testing.T, contract string) CompileInput {
	t.Helper()
	projection, err := ProjectDefinition(fixtureDefinition(t, contract))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := fixtureRegistrySnapshot()
	budget := BudgetProfile{ID: "forms-record-hard-v1", Limits: snapshot.HardLimits}
	published := []PublishedField{
		publishedField(fieldID, "title", TypeText, RoleStored),
		publishedField(countFieldID, "count", TypeInteger, RoleStored),
		publishedField(calcFieldID, "total", TypeInteger, RoleCalculation),
	}
	registrySHA, _ := digestRegistry(snapshot)
	budgetSHA, _ := digestBudget(budget)
	fieldSHA, _ := digestFields(published)
	return CompileInput{Projection: projection,
		Registry: RegistryEvidence{Snapshot: snapshot}, Budget: budget,
		Binding: CatalogBinding{
			TenantID: uuid.MustParse(tenantID), PublicationID: uuid.MustParse(publicationID),
			PublicationRevision: 1, FormID: uuid.MustParse(formID),
			FormVersionID:           uuid.MustParse(versionID),
			SourceDocumentSHA256:    projection.DefinitionSHA256(),
			RecordRulesSHA256:       projection.RecordRulesSHA256(),
			RecordAttestationSHA256: projection.AttestationSHA256(),
			FieldCatalogSHA256:      fieldSHA, BudgetProfileID: budget.ID,
			BudgetSHA256: budgetSHA, RegistrySHA256: registrySHA,
			MappingCatalogSHA256: sha256.Sum256([]byte("mapping-v1")), Fields: published,
		},
	}
}

func publishedField(
	id, name string,
	typeID TypeID,
	role Role,
) PublishedField {
	return PublishedField{FieldID: uuid.MustParse(id), PublicName: name,
		Type: string(typeID), Role: string(role), Generation: 1}
}

func typedLiteral(typeID string, raw string) TypedExpression {
	return TypedExpression{Kind: "literal", Type: typeID,
		Value: json.RawMessage(raw)}
}

func typedGTTotal() TypedExpression {
	operator, field := "gt", calcFieldID
	return TypedExpression{Kind: "binary", Type: "boolean", Operator: &operator,
		Arguments: []TypedExpression{
			{Kind: "field", Type: "integer", FieldID: &field},
			typedLiteral("integer", `"0"`),
		}}
}
