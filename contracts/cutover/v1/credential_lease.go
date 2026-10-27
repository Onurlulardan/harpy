package cutoverv1

type RoleLease struct {
	Principal       string `json:"principal"`
	Login           bool   `json:"login"`
	Replication     bool   `json:"replication"`
	ConnectionLimit int    `json:"connection_limit"`
	PasswordState   string `json:"password_state"`
}

type CredentialLease struct {
	ContractVersion         string            `json:"contract_version"`
	LeaseID                 string            `json:"lease_id"`
	ReleaseID               string            `json:"release_id"`
	EnvironmentID           string            `json:"environment_id"`
	OperationID             string            `json:"operation_id"`
	RunID                   string            `json:"run_id"`
	IssuedAt                string            `json:"issued_at"`
	NotBefore               string            `json:"not_before"`
	ExpiresAt               string            `json:"expires_at"`
	MaxDurationSeconds      int               `json:"max_duration_seconds"`
	ScopeKind               string            `json:"scope_kind"`
	ScopeID                 string            `json:"scope_id"`
	GroupID                 string            `json:"group_id"`
	Domains                 []string          `json:"domains"`
	PlanSHA256              string            `json:"plan_sha256"`
	TransformManifestSHA256 string            `json:"transform_manifest_sha256"`
	Source                  DatabaseIdentity  `json:"source"`
	Target                  DatabaseIdentity  `json:"target"`
	PublicationName         string            `json:"publication_name"`
	SlotName                string            `json:"slot_name"`
	SourceConnectionLimit   int               `json:"source_connection_limit"`
	TargetConnectionLimit   int               `json:"target_connection_limit"`
	RoleLeases              []RoleLease       `json:"role_leases"`
	Relations               []RelationMapping `json:"relations"`
}
