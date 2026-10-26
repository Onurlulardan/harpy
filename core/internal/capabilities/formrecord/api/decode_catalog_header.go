package recordkernel

import "github.com/google/uuid"

func decodeCatalogHeader(
	document catalogAttestationDocument,
	input DecodeCatalogInput,
) (Catalog, error) {
	binding := input.Binding
	if document.Contract != CatalogContract || document.CompilerVersion != CompilerVersion ||
		(document.DefinitionContract != DefinitionV1Contract &&
			document.DefinitionContract != DefinitionV2Contract) ||
		document.BudgetProfileID != binding.BudgetProfileID ||
		document.WorkflowActions == nil || document.Fields == nil ||
		document.Defaults == nil || document.Calculations == nil ||
		document.CrossFieldRules == nil {
		return Catalog{}, ErrCatalogInvalid
	}
	tenantID, tenantErr := parseCatalogTenantID(document.TenantID)
	publicationID, publicationErr := parseCatalogID(document.PublicationID)
	formID, formErr := parseCatalogID(document.FormID)
	versionID, versionErr := parseCatalogID(document.FormVersionID)
	if tenantErr != nil || publicationErr != nil || formErr != nil ||
		versionErr != nil || tenantID != binding.TenantID ||
		publicationID != binding.PublicationID ||
		document.PublicationRevision != binding.PublicationRevision ||
		formID != binding.FormID || versionID != binding.FormVersionID {
		return Catalog{}, ErrCatalogInvalid
	}
	digests, err := decodeCatalogDigests(document)
	if err != nil || digests.definition != binding.SourceDocumentSHA256 ||
		digests.recordRules != binding.RecordRulesSHA256 ||
		digests.recordAttestation != binding.RecordAttestationSHA256 ||
		digests.fields != binding.FieldCatalogSHA256 ||
		digests.budget != binding.BudgetSHA256 ||
		digests.registry != binding.RegistrySHA256 ||
		digests.mapping != binding.MappingCatalogSHA256 {
		return Catalog{}, ErrCatalogInvalid
	}
	limits := document.Limits.limits()
	budget := BudgetProfile{ID: document.BudgetProfileID, Limits: limits}
	budgetSHA, budgetErr := digestBudget(budget)
	fieldSHA, fieldErr := digestFields(binding.Fields)
	if budgetErr != nil || fieldErr != nil || budgetSHA != digests.budget ||
		fieldSHA != digests.fields || !effectiveLimitsValid(limits, expectedHardLimits()) {
		return Catalog{}, ErrCatalogInvalid
	}
	return Catalog{tenantID: tenantID, publicationID: publicationID,
		publicationRevision: document.PublicationRevision, formID: formID,
		formVersionID: versionID, definitionContract: document.DefinitionContract,
		definitionSHA256: digests.definition, fieldCatalogSHA256: digests.fields,
		budgetProfileID: document.BudgetProfileID, budgetSHA256: digests.budget,
		registrySHA256: digests.registry, mappingCatalogSHA256: digests.mapping,
		recordRulesSHA256:    digests.recordRules,
		recordAttestationSHA: digests.recordAttestation,
		limits:               limits}, nil
}

func parseCatalogTenantID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value || !tenantUUID(parsed) {
		return uuid.Nil, ErrCatalogInvalid
	}
	return parsed, nil
}

func parseCatalogID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value || !stableUUIDv7(parsed) {
		return uuid.Nil, ErrCatalogInvalid
	}
	return parsed, nil
}

type catalogDigests struct {
	definition        [32]byte
	fields            [32]byte
	budget            [32]byte
	registry          [32]byte
	mapping           [32]byte
	recordRules       [32]byte
	recordAttestation [32]byte
}

func decodeCatalogDigests(document catalogAttestationDocument) (catalogDigests, error) {
	values := []*[32]byte{}
	result := catalogDigests{}
	values = append(values, &result.definition, &result.fields, &result.budget,
		&result.registry, &result.mapping, &result.recordRules,
		&result.recordAttestation)
	raw := []string{document.DefinitionSHA256, document.FieldCatalogSHA256,
		document.BudgetSHA256, document.RegistrySHA256,
		document.MappingCatalogSHA256, document.RecordRulesSHA256,
		document.RecordAttestationSHA}
	for index := range raw {
		parsed, err := parseHexDigest(raw[index])
		if err != nil {
			return catalogDigests{}, err
		}
		*values[index] = parsed
	}
	return result, nil
}
