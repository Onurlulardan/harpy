package recordkernel

func knownTypeIDs() []TypeID {
	return []TypeID{
		TypeText, TypeInteger, TypeDecimal,
		TypeBoolean, TypeDate, TypeTimestamp,
		TypeUUID, TypeEnum, TypeRelation,
		TypeTextList, TypeUUIDList, TypeJSON,
		TypeTime, TypeDocumentRef,
		TypeDocumentRefList, TypeEnumList,
	}
}

func knownType(typeID TypeID) bool {
	for _, candidate := range knownTypeIDs() {
		if typeID == candidate {
			return true
		}
	}
	return false
}

func expectedHardLimits() BudgetLimits {
	return BudgetLimits{
		DefinitionBytes: 4 << 20, ActiveFields: 500,
		RelationFields: 100, CalculationFields: 100,
		EnumOptionsPerField: 2000, ValidationRules: 2000,
		CalculationASTDepth: 20, CalculationASTNodes: 1000,
		CalculationTotalASTNodes: 20000, DDLIRSteps: 2000,
		PhysicalColumns: 1200, PhysicalIndexes: 128,
		TenantFormTables: 10000, TenantFormIndexes: 50000,
		RecordInputBytes:           MaximumRecordBytes,
		RelationTargetsPerMutation: MaximumRelationTargets,
		CalculationInstructions:    MaximumInstructions,
	}
}
