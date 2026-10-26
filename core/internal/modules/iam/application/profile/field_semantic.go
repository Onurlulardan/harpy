package profile

type fieldSemanticDefinition struct {
	Contract       string               `json:"contract"`
	ID             string               `json:"id"`
	Resource       ResourceKind         `json:"resource_kind"`
	Key            string               `json:"key"`
	Type           LogicalType          `json:"logical_type"`
	Cardinality    Cardinality          `json:"cardinality"`
	Required       bool                 `json:"required"`
	Classification Classification       `json:"classification"`
	Normalization  NormalizationProfile `json:"normalization_profile"`
}

func semanticDefinition(input FieldInput) fieldSemanticDefinition {
	return fieldSemanticDefinition{
		Contract: "harpy/iam-profile-field-semantic/v1",
		ID:       input.ID.String(), Resource: input.Resource, Key: input.Key,
		Type:        input.Type,
		Cardinality: input.Cardinality, Required: input.Required,
		Classification: input.Classification, Normalization: input.Normalization,
	}
}
