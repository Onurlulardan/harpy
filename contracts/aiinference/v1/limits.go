package aiinferencev1

const (
	RequestContractVersion       = "core-ai-inference.v1"
	ResultContractVersion        = "ai-inference-result.v1"
	StreamContractVersion        = "ai-inference-stream-event.v1"
	ReceiptContractVersion       = "ai-inference-usage-receipt.v1"
	SignedReceiptContractVersion = "ai-inference-signed-receipt.v1"
	ServiceAudience              = "harpy-ai-provider-v1"
	ReceiptSigningKeyID          = "ai-core-service-v1"
	ReceiptSigningKeyRevision    = int64(1)

	MaximumRequestBodyBytes    = 1 << 20
	MaximumMessages            = 256
	MaximumEmbeddingInputs     = 128
	MaximumTools               = 64
	MaximumToolProposals       = 64
	MaximumInputTextBytes      = 768 << 10
	MaximumMessageBytes        = 256 << 10
	MaximumToolSchemaBytes     = 64 << 10
	MaximumToolArgumentsBytes  = 64 << 10
	MaximumToolSchemaTotal     = 256 << 10
	MaximumOutputBytes         = 1 << 20
	MaximumEmbeddingDimensions = 8192
	MaximumEmbeddingValues     = 262144
	MinimumOutputTokens        = 1
	MaximumOutputTokens        = 32768
	MaximumReportedTokens      = int64(1_000_000_000)
	MinimumTimeoutMS           = 1000
	MaximumTimeoutMS           = 120000
	StreamWriteTimeoutMS       = 5000
)

const (
	OperationGenerate = "ai.inference.generate"
	OperationStream   = "ai.inference.generate.stream"
	OperationEmbed    = "ai.inference.embed"
	PermissionExecute = "ai.inference.execute"
)

const (
	FinishStop          = "stop"
	FinishLength        = "length"
	FinishToolProposals = "tool_proposals"
	FinishContentFilter = "content_filter"
	FinishEmbedding     = "embedding"
)
