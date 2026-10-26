package runtimev1

type CredentialCleanupReceipt struct {
	ContractVersion string                          `json:"contract_version"`
	Kind            string                          `json:"kind"`
	Payload         CredentialCleanupReceiptPayload `json:"payload"`
	Signature       AttestationSignature            `json:"signature"`
}

type CredentialCleanupReceiptPayload struct {
	RunID                            string             `json:"run_id"`
	ReleaseID                        string             `json:"release_id"`
	IssuedAt                         string             `json:"issued_at"`
	ReleaseManifestSHA256            string             `json:"release_manifest_sha256"`
	BootstrapAuthorityManifestSHA256 string             `json:"bootstrap_authority_manifest_sha256"`
	CleanupOperationDigest           string             `json:"cleanup_operation_digest"`
	DatabaseName                     string             `json:"database_name"`
	DatabaseOrigin                   string             `json:"database_origin"`
	DatabaseSystemIdentifier         string             `json:"database_system_identifier"`
	DatabaseOID                      uint32             `json:"database_oid"`
	DatabaseOwner                    string             `json:"database_owner"`
	DatabaseEncoding                 string             `json:"database_encoding"`
	DatabaseLocaleProvider           string             `json:"database_locale_provider"`
	DatabaseCollation                string             `json:"database_collation"`
	DatabaseCType                    string             `json:"database_ctype"`
	DataChecksums                    bool               `json:"data_checksums"`
	Principals                       []CleanedPrincipal `json:"principals"`
	Result                           string             `json:"result"`
}

type CleanedPrincipal struct {
	Name            string  `json:"name"`
	LeaseID         string  `json:"lease_id"`
	CleanedAt       string  `json:"cleaned_at"`
	CanLogin        bool    `json:"can_login"`
	PasswordState   *string `json:"password_state" harpyNull:"allowed"`
	ActiveSessions  int     `json:"active_sessions"`
	MembershipCount int     `json:"membership_count"`
}
