package tabularmapping

const (
	ContractV2                  = "tabular-mapping.v2"
	FormsImportContractV2       = "forms-record-import-mapping.v2"
	MaximumFields               = 500
	MaximumMappingBytes         = 1024 * 1024
	MaximumRowBytes             = 1024 * 1024
	MaximumSourceBytes          = 256
	MaximumConcatSources        = 32
	MaximumStaticBytes          = 64 * 1024
	MaximumNumericDepth         = 20
	MaximumNumericNodesPerField = 1_000
	MaximumNumericNodes         = 20_000
	MaximumDecimalScale         = 18
	MaximumNumericTextBytes     = 4 * 1024
)
