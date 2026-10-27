package recordkernel

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

func CompileCatalog(input CompileInput) (Catalog, error) {
	binding := input.Binding
	if !binding.valid() || input.Projection.DefinitionSHA256() !=
		sha256.Sum256(input.Projection.CanonicalDefinition()) ||
		input.Projection.DefinitionSHA256() != binding.SourceDocumentSHA256 ||
		input.Projection.RecordRulesSHA256() != binding.RecordRulesSHA256 ||
		input.Projection.AttestationSHA256() != binding.RecordAttestationSHA256 {
		return Catalog{}, ErrCatalogInvalid
	}
	envelope, err := decodeDefinition(input.Projection.StorageProjection())
	if err != nil || envelope.contract != DefinitionV1Contract {
		return Catalog{}, ErrCatalogInvalid
	}
	versionID, err := uuid.Parse(envelope.versionID)
	if err != nil || versionID != binding.FormVersionID {
		return Catalog{}, ErrCatalogInvalid
	}
	var schema StorageSchema
	if strictDecode(envelope.schema, &schema) != nil ||
		schema.FieldRegistryContract != "forms-field-registry.v1" {
		return Catalog{}, ErrCatalogInvalid
	}
	registrySHA, err := digestRegistry(input.Registry.Snapshot)
	if err != nil || registrySHA != binding.RegistrySHA256 {
		return Catalog{}, ErrCatalogInvalid
	}
	budgetSHA, err := digestBudget(input.Budget)
	if err != nil || input.Budget.ID != binding.BudgetProfileID ||
		budgetSHA != binding.BudgetSHA256 ||
		!effectiveLimitsValid(input.Budget.Limits, input.Registry.Snapshot.HardLimits) {
		return Catalog{}, ErrCatalogInvalid
	}
	fieldSHA, err := digestFields(binding.Fields)
	if err != nil || fieldSHA != binding.FieldCatalogSHA256 {
		return Catalog{}, ErrCatalogInvalid
	}
	registry, err := loadRegistry(input.Registry)
	if err != nil {
		return Catalog{}, err
	}
	rules := input.Projection.RecordRules()
	fields, _, _, err := buildFields(schema, registry, binding, rules)
	if err != nil {
		return Catalog{}, fmt.Errorf("%w: fields", err)
	}
	fields, err = applyValidationRules(fields, schema)
	if err != nil {
		return Catalog{}, fmt.Errorf("%w: validations", err)
	}
	byID, byName := indexFields(fields)
	workflowActions, err := compileWorkflowActions(rules.WorkflowActions)
	if err != nil {
		return Catalog{}, fmt.Errorf("%w: workflow actions", err)
	}
	defaults, calculations, crossRules, err := compileRecordExpressions(
		schema, rules, byID, input.Budget.Limits,
	)
	if err != nil {
		return Catalog{}, fmt.Errorf("%w: record expressions", err)
	}
	catalog := Catalog{
		tenantID: binding.TenantID, publicationID: binding.PublicationID,
		publicationRevision: binding.PublicationRevision, formID: binding.FormID,
		formVersionID:      binding.FormVersionID,
		definitionContract: input.Projection.OriginalContract(),
		definitionSHA256:   input.Projection.DefinitionSHA256(),
		fieldCatalogSHA256: binding.FieldCatalogSHA256,
		budgetProfileID:    binding.BudgetProfileID,
		budgetSHA256:       binding.BudgetSHA256, registrySHA256: binding.RegistrySHA256,
		mappingCatalogSHA256: binding.MappingCatalogSHA256,
		recordRulesSHA256:    input.Projection.RecordRulesSHA256(),
		recordAttestationSHA: binding.RecordAttestationSHA256,
		limits:               input.Budget.Limits, fields: fields,
		fieldsByID: byID, fieldsByName: byName, workflowActions: workflowActions,
		defaults: defaults, calculations: calculations,
		crossFieldValidations: crossRules,
	}
	canonical, err := catalogDocument(catalog)
	if err != nil {
		return Catalog{}, fmt.Errorf("%w: attestation", ErrCatalogInvalid)
	}
	catalog.canonical, catalog.sha256 = canonical, sha256.Sum256(canonical)
	return catalog, nil
}

func indexFields(fields []Field) (map[uuid.UUID]Field, map[string]Field) {
	byID := make(map[uuid.UUID]Field, len(fields))
	byName := make(map[string]Field, len(fields))
	for _, field := range fields {
		byID[field.ID], byName[field.PublicName] = field, field
	}
	return byID, byName
}
