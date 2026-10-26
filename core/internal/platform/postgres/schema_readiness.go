package postgres

type SchemaReadinessReason string

const (
	SchemaReady             SchemaReadinessReason = "ready"
	SchemaMissing           SchemaReadinessReason = "missing"
	SchemaFenced            SchemaReadinessReason = "fenced"
	SchemaDatabaseTooOld    SchemaReadinessReason = "database_too_old"
	SchemaRuntimeTooOld     SchemaReadinessReason = "runtime_too_old"
	SchemaManifestMismatch  SchemaReadinessReason = "manifest_mismatch"
	SchemaContractMalformed SchemaReadinessReason = "contract_malformed"
)

type SchemaReadiness struct {
	ready              bool
	reason             SchemaReadinessReason
	databaseGeneration int64
	revision           int64
	verifiedRunID      string
}

func (result SchemaReadiness) Ready() bool                   { return result.ready }
func (result SchemaReadiness) Reason() SchemaReadinessReason { return result.reason }
func (result SchemaReadiness) DatabaseGeneration() int64     { return result.databaseGeneration }
func (result SchemaReadiness) Revision() int64               { return result.revision }
func (result SchemaReadiness) VerifiedRunID() string         { return result.verifiedRunID }
