package fieldregistry

type BudgetLimits struct {
	DefinitionBytes            uint64
	ActiveFields               uint64
	RelationFields             uint64
	CalculationFields          uint64
	EnumOptionsPerField        uint64
	ValidationRules            uint64
	CalculationASTDepth        uint64
	CalculationASTNodes        uint64
	CalculationTotalASTNodes   uint64
	DDLIRSteps                 uint64
	PhysicalColumns            uint64
	PhysicalIndexes            uint64
	TenantFormTables           uint64
	TenantFormIndexes          uint64
	RecordInputBytes           uint64
	RelationTargetsPerMutation uint64
	CalculationInstructions    uint64
}
