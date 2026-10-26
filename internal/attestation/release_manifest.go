package attestation

type ReleaseManifest struct {
	ContractVersion                 string `json:"contract_version"`
	ReleaseID                       string `json:"release_id"`
	IssuedAt                        string `json:"issued_at"`
	ExpiresAt                       string `json:"expires_at"`
	DatabaseName                    string `json:"database_name"`
	DatabaseOrigin                  string `json:"database_origin"`
	DatabaseEncoding                string `json:"database_encoding"`
	DatabaseLocaleProvider          string `json:"database_locale_provider"`
	DatabaseCollation               string `json:"database_collation"`
	DatabaseCType                   string `json:"database_ctype"`
	DatabaseOwner                   string `json:"database_owner"`
	DataChecksums                   bool   `json:"data_checksums"`
	PostgresImageDigest             string `json:"postgres_image_digest"`
	PostgresEngineVersion           string `json:"postgres_engine_version"`
	VectorExtensionVersion          string `json:"vector_extension_version"`
	RuntimeGeneration               int64  `json:"runtime_generation"`
	MinimumDatabaseGeneration       int64  `json:"minimum_database_generation"`
	MigrationManifestSHA256         string `json:"migration_manifest_sha256"`
	ExtensionManifestSHA256         string `json:"extension_manifest_sha256"`
	RolesManifestSHA256             string `json:"roles_manifest_sha256"`
	PrivilegesManifestSHA256        string `json:"privileges_manifest_sha256"`
	OwnershipManifestSHA256         string `json:"ownership_manifest_sha256"`
	BootstrapAuthoritySHA256        string `json:"bootstrap_authority_manifest_sha256"`
	BootstrapOperationDigest        string `json:"bootstrap_operation_digest"`
	ExtensionOperationDigest        string `json:"extension_operation_digest"`
	CredentialCleanupDigest         string `json:"credential_cleanup_operation_digest"`
	CutoverAuthorityOperationDigest string `json:"cutover_authority_adoption_operation_digest"`
	CatalogPackageSetSHA256         string `json:"catalog_package_set_sha256"`
}
