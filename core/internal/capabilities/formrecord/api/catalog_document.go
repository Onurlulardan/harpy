package recordkernel

type catalogFieldDocument struct {
	ID          string           `json:"id"`
	PublicName  string           `json:"public_name"`
	Type        string           `json:"type"`
	Role        string           `json:"role"`
	Profile     string           `json:"profile"`
	Nullable    bool             `json:"nullable"`
	WritePolicy WritePolicy      `json:"write_policy"`
	Generation  uint32           `json:"generation"`
	Decimal     *DecimalProfile  `json:"decimal,omitempty"`
	EnumOptions []string         `json:"enum_options,omitempty"`
	Relation    *RelationProfile `json:"relation,omitempty"`
	Required    bool             `json:"required"`
	Immutable   bool             `json:"immutable"`
	TextMinimum *int             `json:"text_minimum,omitempty"`
	TextMaximum *int             `json:"text_maximum,omitempty"`
}

func catalogDocument(catalog Catalog) ([]byte, error) {
	fields := make([]catalogFieldDocument, len(catalog.fields))
	for index, field := range catalog.fields {
		fields[index] = catalogFieldDocument{
			ID: field.ID.String(), PublicName: field.PublicName,
			Type: string(field.TypeID), Role: string(field.Role),
			Profile: string(field.ValueProfile), Nullable: field.Nullable,
			WritePolicy: field.WritePolicy, Generation: field.Generation,
			Decimal: field.Decimal, EnumOptions: append([]string(nil), field.EnumOptions...),
			Relation: field.Relation, Required: field.Required,
			Immutable: field.Immutable, TextMinimum: field.TextMinimum,
			TextMaximum: field.TextMaximum,
		}
	}
	defaults, err := defaultDocuments(catalog.defaults)
	if err != nil {
		return nil, err
	}
	calculations, err := calculationDocuments(catalog.calculations)
	if err != nil {
		return nil, err
	}
	crossRules, err := crossRuleDocuments(catalog.crossFieldValidations)
	if err != nil {
		return nil, err
	}
	document := catalogAttestationDocument{
		Contract: CatalogContract, CompilerVersion: CompilerVersion,
		TenantID: catalog.tenantID.String(), PublicationID: catalog.publicationID.String(),
		PublicationRevision: catalog.publicationRevision, FormID: catalog.formID.String(),
		FormVersionID:        catalog.formVersionID.String(),
		DefinitionContract:   catalog.definitionContract,
		DefinitionSHA256:     hexDigest(catalog.definitionSHA256),
		FieldCatalogSHA256:   hexDigest(catalog.fieldCatalogSHA256),
		BudgetProfileID:      catalog.budgetProfileID,
		BudgetSHA256:         hexDigest(catalog.budgetSHA256),
		RegistrySHA256:       hexDigest(catalog.registrySHA256),
		MappingCatalogSHA256: hexDigest(catalog.mappingCatalogSHA256),
		RecordRulesSHA256:    hexDigest(catalog.recordRulesSHA256),
		RecordAttestationSHA: hexDigest(catalog.recordAttestationSHA),
		Limits:               catalogLimits(catalog.limits),
		WorkflowActions:      workflowActionStrings(catalog), Fields: fields,
		Defaults: defaults, Calculations: calculations, CrossFieldRules: crossRules,
	}
	return canonicalObject(document, MaximumCatalogBytes)
}

func workflowActionStrings(catalog Catalog) []string {
	actions := catalog.WorkflowActions()
	result := make([]string, len(actions))
	for index, action := range actions {
		result[index] = string(action)
	}
	return result
}
