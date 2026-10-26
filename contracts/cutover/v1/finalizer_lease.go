package cutoverv1

type FinalizerLease struct {
	ContractVersion                string           `json:"contract_version"`
	LeaseID                        string           `json:"lease_id"`
	ReleaseID                      string           `json:"release_id"`
	EnvironmentID                  string           `json:"environment_id"`
	OperationID                    string           `json:"operation_id"`
	RunID                          string           `json:"run_id"`
	IssuedAt                       string           `json:"issued_at"`
	NotBefore                      string           `json:"not_before"`
	ExpiresAt                      string           `json:"expires_at"`
	MaxDurationSeconds             int              `json:"max_duration_seconds"`
	RoleLease                      RoleLease        `json:"role_lease"`
	Target                         DatabaseIdentity `json:"target"`
	AdoptionLeaseSHA256            string           `json:"adoption_lease_sha256"`
	CleanupReceiptSHA256           string           `json:"cleanup_receipt_sha256"`
	AuthorityAdoptionReceiptSHA256 string           `json:"authority_adoption_receipt_sha256"`
	FinalizationStateSHA256        string           `json:"finalization_state_sha256"`
	TenantRequirementCatalogSHA256 *string          `json:"tenant_requirement_catalog_sha256" harpyNull:"allowed"`
	TenantRequirementCount         *int             `json:"tenant_requirement_count" harpyNull:"allowed"`
	FinalizationProcedureOID       uint32           `json:"finalization_procedure_oid"`
	FinalizationProcedureSignature string           `json:"finalization_procedure_signature"`
	TerminalState                  string           `json:"terminal_state"`
	TerminalEvidenceSHA256         string           `json:"terminal_evidence_sha256"`
}

type FinalizerLeaseSignature struct {
	ContractVersion string    `json:"contract_version"`
	ManifestSHA256  string    `json:"manifest_sha256"`
	Signature       Signature `json:"signature"`
}
