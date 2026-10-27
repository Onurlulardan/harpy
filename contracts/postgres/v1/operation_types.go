package postgresv1

type OperationStep struct {
	StepID           string `json:"step_id"`
	TransactionClass string `json:"transaction_class"`
}

type BootstrapOperation struct {
	ContractVersion string               `json:"contract_version"`
	DatabaseOrigin  string               `json:"database_origin"`
	DatabaseName    string               `json:"database_name"`
	InputHashes     BootstrapInputHashes `json:"input_hashes"`
	Steps           []OperationStep      `json:"steps"`
}

type BootstrapInputHashes struct {
	RolesManifestSHA256      string `json:"roles_manifest_sha256"`
	PrivilegesManifestSHA256 string `json:"privileges_manifest_sha256"`
	OwnershipManifestSHA256  string `json:"ownership_manifest_sha256"`
}

type ExtensionOperation struct {
	ContractVersion string               `json:"contract_version"`
	DatabaseOrigin  string               `json:"database_origin"`
	DatabaseName    string               `json:"database_name"`
	InputHashes     ExtensionInputHashes `json:"input_hashes"`
	Steps           []OperationStep      `json:"steps"`
}

type ExtensionInputHashes struct {
	ExtensionManifestSHA256 string `json:"extension_manifest_sha256"`
	OwnershipManifestSHA256 string `json:"ownership_manifest_sha256"`
}

type CredentialCleanupOperation struct {
	ContractVersion string                  `json:"contract_version"`
	DatabaseOrigin  string                  `json:"database_origin"`
	DatabaseName    string                  `json:"database_name"`
	InputHashes     CredentialCleanupHashes `json:"input_hashes"`
	PrivilegedRoles []string                `json:"privileged_roles"`
	Steps           []OperationStep         `json:"steps"`
}

type CredentialCleanupHashes struct {
	RolesManifestSHA256      string `json:"roles_manifest_sha256"`
	PrivilegesManifestSHA256 string `json:"privileges_manifest_sha256"`
}
