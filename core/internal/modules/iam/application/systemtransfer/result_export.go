package systemtransfer

type ExportExecutionState string

const (
	ExportExecutionReady          ExportExecutionState = "ready"
	ExportExecutionCancelled      ExportExecutionState = "cancelled"
	ExportExecutionRetry          ExportExecutionState = "retry"
	ExportExecutionFailed         ExportExecutionState = "failed"
	ExportExecutionLeaseLost      ExportExecutionState = "lease_lost"
	ExportExecutionOutcomeUnknown ExportExecutionState = "outcome_unknown"
)

const (
	ExportCodeAuthorizationRevoked  = "IAM_SYSTEM_EXPORT_AUTH_REVOKED"
	ExportCodeCancelled             = "IAM_SYSTEM_EXPORT_CANCELLED"
	ExportCodeLimitExceeded         = "IAM_SYSTEM_EXPORT_LIMIT_EXCEEDED"
	ExportCodeQueryInvalid          = "IAM_SYSTEM_EXPORT_QUERY_INVALID"
	ExportCodeCatalogPinStale       = "IAM_SYSTEM_EXPORT_CATALOG_PIN_STALE"
	ExportCodeArtifactTampered      = "IAM_SYSTEM_EXPORT_ARTIFACT_TAMPERED"
	ExportCodeStateConflict         = "IAM_SYSTEM_EXPORT_STATE_CONFLICT"
	ExportCodeDependencyUnavailable = "IAM_SYSTEM_EXPORT_DEPENDENCY_UNAVAILABLE"
	ExportCodeOutcomeUnknown        = ExportOutcomeUnknownCode
)

type ExportExecutionResult struct {
	State ExportExecutionState
	Code  string
}
