package runtimev1

type ExtensionReceipt struct {
	ContractVersion string                  `json:"contract_version"`
	Kind            string                  `json:"kind"`
	Payload         ExtensionReceiptPayload `json:"payload"`
	Signature       AttestationSignature    `json:"signature"`
}

type ExtensionReceiptPayload struct {
	RunID                            string            `json:"run_id"`
	ReleaseID                        string            `json:"release_id"`
	IssuedAt                         string            `json:"issued_at"`
	ReleaseManifestSHA256            string            `json:"release_manifest_sha256"`
	BootstrapAuthorityManifestSHA256 string            `json:"bootstrap_authority_manifest_sha256"`
	CredentialLeaseID                string            `json:"credential_lease_id"`
	PostgresImageDigest              string            `json:"postgres_image_digest"`
	OwnershipManifestSHA256          string            `json:"ownership_manifest_sha256"`
	DatabaseName                     string            `json:"database_name"`
	DatabaseSystemIdentifier         string            `json:"database_system_identifier"`
	DatabaseOID                      uint32            `json:"database_oid"`
	ServerVersionNumber              int               `json:"server_version_num"`
	DatabaseOrigin                   string            `json:"database_origin"`
	DatabaseOwner                    string            `json:"database_owner"`
	DatabaseEncoding                 string            `json:"database_encoding"`
	DatabaseLocaleProvider           string            `json:"database_locale_provider"`
	DatabaseCollation                string            `json:"database_collation"`
	DatabaseCType                    string            `json:"database_ctype"`
	DataChecksums                    bool              `json:"data_checksums"`
	SessionUser                      string            `json:"session_user"`
	CurrentUser                      string            `json:"current_user"`
	OperationDigest                  string            `json:"operation_digest"`
	Result                           string            `json:"result"`
	ExtensionManifestSHA256          string            `json:"extension_manifest_sha256"`
	Extensions                       []ExtensionRecord `json:"extensions"`
}

type ExtensionRecord struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	Schema            string `json:"schema"`
	Owner             string `json:"owner"`
	ControlFileSHA256 string `json:"control_file_sha256"`
}
