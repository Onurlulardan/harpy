package runtimev1

type BootstrapReceipt struct {
	ContractVersion string                  `json:"contract_version"`
	Kind            string                  `json:"kind"`
	Payload         BootstrapReceiptPayload `json:"payload"`
	Signature       AttestationSignature    `json:"signature"`
}

type BootstrapReceiptPayload struct {
	RunID                            string `json:"run_id"`
	ReleaseID                        string `json:"release_id"`
	EnvironmentID                    string `json:"environment_id"`
	IssuedAt                         string `json:"issued_at"`
	ReleaseManifestSHA256            string `json:"release_manifest_sha256"`
	BootstrapAuthorityManifestSHA256 string `json:"bootstrap_authority_manifest_sha256"`
	CredentialLeaseID                string `json:"credential_lease_id"`
	PostgresImageDigest              string `json:"postgres_image_digest"`
	OwnershipManifestSHA256          string `json:"ownership_manifest_sha256"`
	DatabaseName                     string `json:"database_name"`
	DatabaseSystemIdentifier         string `json:"database_system_identifier"`
	DatabaseOID                      uint32 `json:"database_oid"`
	ServerVersionNumber              int    `json:"server_version_num"`
	DatabaseOrigin                   string `json:"database_origin"`
	DatabaseOwner                    string `json:"database_owner"`
	DatabaseEncoding                 string `json:"database_encoding"`
	DatabaseLocaleProvider           string `json:"database_locale_provider"`
	DatabaseCollation                string `json:"database_collation"`
	DatabaseCType                    string `json:"database_ctype"`
	DataChecksums                    bool   `json:"data_checksums"`
	SessionUser                      string `json:"session_user"`
	CurrentUser                      string `json:"current_user"`
	OperationDigest                  string `json:"operation_digest"`
	Result                           string `json:"result"`
	RolesManifestSHA256              string `json:"roles_manifest_sha256"`
	PrivilegesManifestSHA256         string `json:"privileges_manifest_sha256"`
}
