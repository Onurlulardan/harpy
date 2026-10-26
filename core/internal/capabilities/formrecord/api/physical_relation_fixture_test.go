package recordkernel

import (
	"encoding/json"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const (
	relationFieldID       = "018f47f4-7b1b-7cc0-8000-000000001011"
	relationID            = "018f47f4-7b1b-7cc0-8000-000000001012"
	externalFormID        = "018f47f4-7b1b-7cc0-8000-000000001013"
	externalPublicationID = "018f47f4-7b1b-7cc0-8000-000000001014"
	externalFormVersionID = "018f47f4-7b1b-7cc0-8000-000000001015"
	relationSourceGuardID = "018f47f4-7b1b-7cc0-8000-000000001016"
	relationTargetGuardID = "018f47f4-7b1b-7cc0-8000-000000001017"
)

func relationPhysicalFixture(
	t *testing.T, targetFormID string,
) (Catalog, []PhysicalMapping, [32]byte) {
	t.Helper()
	projection, err := ProjectDefinition(relationDefinitionFixture(t, targetFormID))
	if err != nil {
		t.Fatal(err)
	}
	registry := fixtureRegistrySnapshot()
	budget := BudgetProfile{ID: "forms-record-hard-v1", Limits: registry.HardLimits}
	published := []PublishedField{publishedField(
		relationFieldID, "links", TypeRelation, RoleRelation,
	)}
	registrySHA, _ := digestRegistry(registry)
	budgetSHA, _ := digestBudget(budget)
	fieldSHA, _ := digestFields(published)
	mappings := []PhysicalMapping{
		physicalObject(t, formID, physicalFormTable, "record", "forms_record_table_v2"),
		physicalObject(t, relationID, physicalRelationTable, "relation", "forms_relation_edge_v1"),
	}
	mappingSHA, err := DigestMappings(mappings)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := CompileCatalog(CompileInput{Projection: projection,
		Registry: RegistryEvidence{Snapshot: registry}, Budget: budget,
		Binding: CatalogBinding{
			TenantID: uuid.MustParse(tenantID), PublicationID: uuid.MustParse(publicationID),
			PublicationRevision: 1, FormID: uuid.MustParse(formID),
			FormVersionID: uuid.MustParse(versionID), SourceDocumentSHA256: projection.DefinitionSHA256(),
			RecordRulesSHA256:       projection.RecordRulesSHA256(),
			RecordAttestationSHA256: projection.AttestationSHA256(),
			FieldCatalogSHA256:      fieldSHA, BudgetProfileID: budget.ID,
			BudgetSHA256: budgetSHA, RegistrySHA256: registrySHA,
			MappingCatalogSHA256: mappingSHA, Fields: published,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return catalog, mappings, mappingSHA
}

func relationDefinitionFixture(t *testing.T, targetFormID string) []byte {
	t.Helper()
	schema := StorageSchema{FieldRegistryContract: RegistryContractVersion,
		Fields: []FieldDefinition{{ID: relationFieldID, Name: "links",
			TypeID: TypeRelation, Role: RoleRelation, Nullable: true}},
		Relations: []RelationDefinition{{ID: relationID, FieldID: relationFieldID,
			TargetFormID: targetFormID, SourceConstraintID: relationSourceGuardID,
			TargetConstraintID: relationTargetGuardID, Cardinality: "many",
			DeletePolicy: "restrict"}},
		Indexes: []IndexDefinition{}, Constraints: []ConstraintDefinition{},
		Validations: []ValidationDefinition{}, Calculations: []CalculationDefinition{}}
	wire := map[string]any{"contract_version": DefinitionV1Contract,
		"form_version_id": versionID, "schema_contract": StorageSchemaContractVersion,
		"schema": schema, "presentation_contract": PresentationContractVersion,
		"presentation": map[string]any{"layout": "default"}}
	encoded, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, MaximumCatalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
