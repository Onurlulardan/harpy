package runtimev1

type BootstrapAuthorityManifest struct {
	ContractVersion                  string                         `json:"contract_version"`
	ManifestID                       string                         `json:"manifest_id"`
	EnvironmentID                    string                         `json:"environment_id"`
	ReleaseID                        string                         `json:"release_id"`
	IssuedAt                         string                         `json:"issued_at"`
	ExpiresAt                        string                         `json:"expires_at"`
	DatabaseOrigin                   string                         `json:"database_origin"`
	Database                         BootstrapAuthorityDatabase     `json:"database"`
	Connections                      BootstrapAuthorityConnections  `json:"connections"`
	CreateRoleSelfGrant              string                         `json:"createrole_self_grant"`
	Roles                            []BootstrapAuthorityRole       `json:"roles"`
	Memberships                      []BootstrapAuthorityMembership `json:"memberships"`
	CredentialLeases                 []BootstrapCredentialLease     `json:"credential_leases"`
	RolesManifestSHA256              string                         `json:"roles_manifest_sha256"`
	PrivilegesManifestSHA256         string                         `json:"privileges_manifest_sha256"`
	OwnershipManifestSHA256          string                         `json:"ownership_manifest_sha256"`
	BootstrapOperationDigest         string                         `json:"bootstrap_operation_digest"`
	CredentialCleanupOperationDigest string                         `json:"credential_cleanup_operation_digest"`
}

type BootstrapAuthorityDatabase struct {
	Name                           string  `json:"name"`
	CreateDatabase                 bool    `json:"create_database"`
	SourceOwner                    *string `json:"source_owner" harpyNull:"allowed"`
	Owner                          string  `json:"owner"`
	ExtensionOwner                 string  `json:"extension_owner"`
	ExtensionOwnerMode             string  `json:"extension_owner_mode"`
	ExtensionOwnerApplicationLogin bool    `json:"extension_owner_application_login"`
	OwnershipExecutor              string  `json:"ownership_executor"`
	OwnershipAttestation           string  `json:"ownership_attestation"`
	Encoding                       string  `json:"encoding"`
	LocaleProvider                 string  `json:"locale_provider"`
	Collation                      string  `json:"collation"`
	CType                          string  `json:"ctype"`
	DataChecksums                  bool    `json:"data_checksums"`
	Template                       *string `json:"template" harpyNull:"allowed"`
	PublicConnect                  bool    `json:"public_connect"`
	PublicCreate                   bool    `json:"public_create"`
	PublicTemp                     bool    `json:"public_temp"`
}

type BootstrapAuthorityConnections struct {
	EndpointKind        string `json:"endpoint_kind"`
	MaintenanceDatabase string `json:"maintenance_database"`
	TargetDatabase      string `json:"target_database"`
	SameServerRequired  bool   `json:"same_server_required"`
}

type BootstrapAuthorityRole struct {
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	CredentialMode  string `json:"credential_mode"`
	Login           bool   `json:"login"`
	Inherit         bool   `json:"inherit"`
	Superuser       bool   `json:"superuser"`
	CreateDB        bool   `json:"createdb"`
	CreateRole      bool   `json:"createrole"`
	Replication     bool   `json:"replication"`
	BypassRLS       bool   `json:"bypassrls"`
	ConnectionLimit int    `json:"connection_limit"`
}

type BootstrapAuthorityMembership struct {
	Member  string `json:"member"`
	Role    string `json:"role"`
	Enabled bool   `json:"enabled"`
	Admin   bool   `json:"admin"`
	Inherit bool   `json:"inherit"`
	Set     bool   `json:"set"`
	Grantor string `json:"grantor"`
}

type BootstrapCredentialLease struct {
	LeaseID            string `json:"lease_id"`
	Principal          string `json:"principal"`
	NotBefore          string `json:"not_before"`
	ExpiresAt          string `json:"expires_at"`
	MaxDurationSeconds int    `json:"max_duration_seconds"`
	MaxConnections     int    `json:"max_connections"`
	CleanupPolicy      string `json:"cleanup_policy"`
}

type BootstrapAuthoritySignature struct {
	ContractVersion string               `json:"contract_version"`
	ManifestSHA256  string               `json:"manifest_sha256"`
	Signature       AttestationSignature `json:"signature"`
}
