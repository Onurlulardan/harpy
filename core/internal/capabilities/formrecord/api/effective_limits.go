package recordkernel

func effectiveLimitsValid(value, hard BudgetLimits) bool {
	values := []uint64{
		value.DefinitionBytes, value.ActiveFields, value.RelationFields,
		value.CalculationFields, value.EnumOptionsPerField, value.ValidationRules,
		value.CalculationASTDepth, value.CalculationASTNodes,
		value.CalculationTotalASTNodes, value.DDLIRSteps, value.PhysicalColumns,
		value.PhysicalIndexes, value.TenantFormTables, value.TenantFormIndexes,
		value.RecordInputBytes, value.RelationTargetsPerMutation,
		value.CalculationInstructions,
	}
	hards := []uint64{
		hard.DefinitionBytes, hard.ActiveFields, hard.RelationFields,
		hard.CalculationFields, hard.EnumOptionsPerField, hard.ValidationRules,
		hard.CalculationASTDepth, hard.CalculationASTNodes,
		hard.CalculationTotalASTNodes, hard.DDLIRSteps, hard.PhysicalColumns,
		hard.PhysicalIndexes, hard.TenantFormTables, hard.TenantFormIndexes,
		hard.RecordInputBytes, hard.RelationTargetsPerMutation,
		hard.CalculationInstructions,
	}
	for index, limit := range values {
		if limit < 1 || limit > hards[index] {
			return false
		}
	}
	return true
}
