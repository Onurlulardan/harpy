package recordkernel

type catalogLimitsDocument struct {
	DefinitionBytes            uint64 `json:"definition_bytes"`
	ActiveFields               uint64 `json:"active_fields"`
	RelationFields             uint64 `json:"relation_fields"`
	CalculationFields          uint64 `json:"calculation_fields"`
	EnumOptionsPerField        uint64 `json:"enum_options_per_field"`
	ValidationRules            uint64 `json:"validation_rules"`
	CalculationASTDepth        uint64 `json:"calculation_ast_depth"`
	CalculationASTNodes        uint64 `json:"calculation_ast_nodes"`
	CalculationTotalASTNodes   uint64 `json:"calculation_total_ast_nodes"`
	DDLIRSteps                 uint64 `json:"ddl_ir_steps"`
	PhysicalColumns            uint64 `json:"physical_columns"`
	PhysicalIndexes            uint64 `json:"physical_indexes"`
	TenantFormTables           uint64 `json:"tenant_form_tables"`
	TenantFormIndexes          uint64 `json:"tenant_form_indexes"`
	RecordInputBytes           uint64 `json:"record_input_bytes"`
	RelationTargetsPerMutation uint64 `json:"relation_targets_per_mutation"`
	CalculationInstructions    uint64 `json:"calculation_instructions"`
}

func catalogLimits(value BudgetLimits) catalogLimitsDocument {
	return catalogLimitsDocument{
		value.DefinitionBytes, value.ActiveFields, value.RelationFields,
		value.CalculationFields, value.EnumOptionsPerField, value.ValidationRules,
		value.CalculationASTDepth, value.CalculationASTNodes,
		value.CalculationTotalASTNodes, value.DDLIRSteps, value.PhysicalColumns,
		value.PhysicalIndexes, value.TenantFormTables, value.TenantFormIndexes,
		value.RecordInputBytes, value.RelationTargetsPerMutation,
		value.CalculationInstructions,
	}
}

func (value catalogLimitsDocument) limits() BudgetLimits {
	return BudgetLimits{
		DefinitionBytes: value.DefinitionBytes, ActiveFields: value.ActiveFields,
		RelationFields: value.RelationFields, CalculationFields: value.CalculationFields,
		EnumOptionsPerField:      value.EnumOptionsPerField,
		ValidationRules:          value.ValidationRules,
		CalculationASTDepth:      value.CalculationASTDepth,
		CalculationASTNodes:      value.CalculationASTNodes,
		CalculationTotalASTNodes: value.CalculationTotalASTNodes,
		DDLIRSteps:               value.DDLIRSteps, PhysicalColumns: value.PhysicalColumns,
		PhysicalIndexes: value.PhysicalIndexes, TenantFormTables: value.TenantFormTables,
		TenantFormIndexes:          value.TenantFormIndexes,
		RecordInputBytes:           value.RecordInputBytes,
		RelationTargetsPerMutation: value.RelationTargetsPerMutation,
		CalculationInstructions:    value.CalculationInstructions,
	}
}
