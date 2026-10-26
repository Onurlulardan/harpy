package recordkernel

type catalogAttestationDocument struct {
	Contract             string                 `json:"contract"`
	CompilerVersion      string                 `json:"compiler_version"`
	TenantID             string                 `json:"tenant_id"`
	PublicationID        string                 `json:"publication_id"`
	PublicationRevision  int64                  `json:"publication_revision"`
	FormID               string                 `json:"form_id"`
	FormVersionID        string                 `json:"form_version_id"`
	DefinitionContract   string                 `json:"definition_contract"`
	DefinitionSHA256     string                 `json:"definition_sha256"`
	FieldCatalogSHA256   string                 `json:"field_catalog_sha256"`
	BudgetProfileID      string                 `json:"budget_profile_id"`
	BudgetSHA256         string                 `json:"budget_sha256"`
	RegistrySHA256       string                 `json:"registry_sha256"`
	MappingCatalogSHA256 string                 `json:"mapping_catalog_sha256"`
	RecordRulesSHA256    string                 `json:"record_rules_sha256"`
	RecordAttestationSHA string                 `json:"record_attestation_sha256"`
	Limits               catalogLimitsDocument  `json:"limits"`
	WorkflowActions      []string               `json:"workflow_actions"`
	Fields               []catalogFieldDocument `json:"fields"`
	Defaults             []catalogAssignment    `json:"defaults"`
	Calculations         []catalogAssignment    `json:"calculations"`
	CrossFieldRules      []catalogCrossRule     `json:"cross_field_validations"`
}
