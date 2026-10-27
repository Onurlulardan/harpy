package recordkernel

const (
	DefinitionV1Contract = "forms-definition.v1"
	DefinitionV2Contract = "forms-definition.v2"
	RecordRulesContract  = "forms-record-rules.v1"
	CatalogContract      = "forms-record-catalog.v1"
	ProjectionContract   = "forms-definition-projection.v1"
	CompilerVersion      = "1.0.0"
)

const (
	MaximumRecordBytes         = 1 << 20
	MaximumCatalogBytes        = 4 << 20
	MaximumFieldBytes          = 1 << 20
	MaximumListElements        = 10_000
	MaximumJSONDepth           = 128
	MaximumRelationEdges       = 100
	MaximumRelationTargets     = 1_000
	MaximumInstructions        = 100_000
	MaximumExpressionDepth     = 20
	MaximumExpressionNodes     = 1_000
	MaximumTotalASTNodes       = 20_000
	MaximumCrossFieldRules     = 2_000
	MaximumExpressionArguments = 16
)
