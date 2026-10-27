package fieldregistry

const (
	CostScalar            CostClass = "scalar"
	CostNumeric           CostClass = "numeric"
	CostTextProfiled      CostClass = "text_profiled"
	CostCollection        CostClass = "collection"
	CostJSON              CostClass = "json"
	CostRelationEdge      CostClass = "relation_edge"
	CostDocumentReference CostClass = "document_reference"
)

const (
	BudgetDefinitionBytes          BudgetDimension = "definition_bytes"
	BudgetActiveFields             BudgetDimension = "active_fields"
	BudgetRelationFields           BudgetDimension = "relation_fields"
	BudgetCalculationFields        BudgetDimension = "calculation_fields"
	BudgetEnumOptions              BudgetDimension = "enum_options"
	BudgetValidationRules          BudgetDimension = "validation_rules"
	BudgetCalculationASTDepth      BudgetDimension = "calculation_ast_depth"
	BudgetCalculationASTNodes      BudgetDimension = "calculation_ast_nodes"
	BudgetCalculationTotalASTNodes BudgetDimension = "calculation_total_ast_nodes"
	BudgetDDLIRSteps               BudgetDimension = "ddl_ir_steps"
	BudgetPhysicalColumns          BudgetDimension = "physical_columns"
	BudgetPhysicalIndexes          BudgetDimension = "physical_indexes"
	BudgetTenantFormTables         BudgetDimension = "tenant_form_tables"
	BudgetTenantFormIndexes        BudgetDimension = "tenant_form_indexes"
	BudgetRecordInputBytes         BudgetDimension = "record_input_bytes"
	BudgetRelationTargets          BudgetDimension = "relation_targets"
	BudgetCalculationInstructions  BudgetDimension = "calculation_instructions"
)
