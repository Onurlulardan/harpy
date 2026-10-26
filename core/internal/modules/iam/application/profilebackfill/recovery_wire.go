package profilebackfill

type recoveryPlanWire struct {
	ID                            string        `json:"id"`
	ResourceKind                  string        `json:"resource_kind"`
	RecoveryOfPlanID              string        `json:"recovery_of_plan_id"`
	RecoveryRootPlanID            string        `json:"recovery_root_plan_id"`
	RecoverySequence              int64         `json:"recovery_sequence"`
	ParentPlanSHA256              string        `json:"parent_plan_sha256"`
	SourceSchemaID                string        `json:"source_schema_id"`
	TargetSchemaID                string        `json:"target_schema_id"`
	TargetVersion                 int64         `json:"target_version"`
	TargetDefinitionSHA256        string        `json:"target_definition_sha256"`
	ConverterVersion              string        `json:"converter_version"`
	InheritedTargetCount          int64         `json:"inherited_target_count"`
	InheritedTargetSHA256         string        `json:"inherited_target_sha256"`
	RemainingSourceCount          int64         `json:"remaining_source_count"`
	RemainingSourceSHA256         string        `json:"remaining_source_sha256"`
	ExpectedRemainingTargetSHA256 string        `json:"expected_remaining_target_sha256"`
	PageSize                      int           `json:"page_size"`
	InitialRollingSHA256          string        `json:"initial_rolling_sha256"`
	Mappings                      []mappingWire `json:"mappings"`
	Pages                         []pageWire    `json:"pages"`
}
