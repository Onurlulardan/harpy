package cutoverv1

type AuthorityAdoptionObject struct {
	Kind                    string `json:"kind"`
	SchemaName              string `json:"schema_name"`
	ObjectName              string `json:"object_name"`
	IdentityArguments       string `json:"identity_arguments"`
	Action                  string `json:"action"`
	IntroducedGeneration    int    `json:"introduced_generation"`
	PreAdoptionOwner        string `json:"pre_adoption_owner"`
	AdoptedOwner            string `json:"adopted_owner"`
	PreDefinitionSHA256     string `json:"pre_definition_sha256"`
	AdoptedDefinitionSHA256 string `json:"adopted_definition_sha256"`
	PreAdoptionACLSHA256    string `json:"pre_adoption_acl_sha256"`
	AdoptedACLSHA256        string `json:"adopted_acl_sha256"`
}

type AuthorityCatalogDependency struct {
	Kind              string `json:"kind"`
	SchemaName        string `json:"schema_name"`
	ObjectName        string `json:"object_name"`
	IdentityArguments string `json:"identity_arguments"`
	Owner             string `json:"owner"`
	DefinitionSHA256  string `json:"definition_sha256"`
	ACLSHA256         string `json:"acl_sha256"`
}

type AuthorityAdoptionOperation struct {
	ContractVersion                string                       `json:"contract_version"`
	OperationID                    string                       `json:"operation_id"`
	ReleaseID                      string                       `json:"release_id"`
	EnvironmentID                  string                       `json:"environment_id"`
	Mode                           string                       `json:"mode"`
	PreviousCatalogGeneration      int                          `json:"previous_catalog_generation"`
	CatalogGeneration              int                          `json:"catalog_generation"`
	PreCatalogContractSHA256       string                       `json:"pre_catalog_contract_sha256"`
	PostCatalogContractSHA256      string                       `json:"post_catalog_contract_sha256"`
	PreviousAuthorityReceiptSHA256 *string                      `json:"previous_authority_receipt_sha256" harpyNull:"allowed"`
	UpgradeArtifactSHA256          *string                      `json:"upgrade_artifact_sha256" harpyNull:"allowed"`
	IssuedAt                       string                       `json:"issued_at"`
	ExpiresAt                      string                       `json:"expires_at"`
	Target                         DatabaseIdentity             `json:"target"`
	MigrationManifestSHA256        string                       `json:"migration_manifest_sha256"`
	BootstrapAuthoritySHA256       string                       `json:"bootstrap_authority_manifest_sha256"`
	Executor                       string                       `json:"executor"`
	ExecutorSessionUser            string                       `json:"executor_session_user"`
	OwnershipMode                  string                       `json:"ownership_mode"`
	MembershipPolicy               string                       `json:"membership_policy"`
	QuiescencePolicy               AuthorityQuiescencePolicy    `json:"quiescence_policy"`
	AuthorityOwnerRole             AuthorityOwnerRoleState      `json:"authority_owner_role"`
	Objects                        []AuthorityAdoptionObject    `json:"objects"`
	DependencyContractSHA256       string                       `json:"dependency_contract_sha256"`
	Dependencies                   []AuthorityCatalogDependency `json:"dependencies"`
}

type AuthorityOwnerRoleState struct {
	Login            bool   `json:"login"`
	Inherit          bool   `json:"inherit"`
	Superuser        bool   `json:"superuser"`
	CreateDB         bool   `json:"createdb"`
	CreateRole       bool   `json:"createrole"`
	Replication      bool   `json:"replication"`
	BypassRLS        bool   `json:"bypassrls"`
	ConnectionLimit  int    `json:"connection_limit"`
	PasswordState    string `json:"password_state"`
	ValidUntil       string `json:"valid_until"`
	MembershipCount  int    `json:"membership_count"`
	OwnedObjectCount int    `json:"owned_object_count"`
}

type AuthorityAdoptionReceiptObject struct {
	Kind                 string `json:"kind"`
	SchemaName           string `json:"schema_name"`
	ObjectName           string `json:"object_name"`
	IdentityArguments    string `json:"identity_arguments"`
	IntroducedGeneration int    `json:"introduced_generation"`
	ObjectOID            uint32 `json:"object_oid"`
	Owner                string `json:"owner"`
	DefinitionSHA256     string `json:"definition_sha256"`
	ACLSHA256            string `json:"acl_sha256"`
}

type AuthorityCatalogDependencyReceipt struct {
	AuthorityCatalogDependency
	ObjectOID uint32 `json:"object_oid"`
}

type AuthorityCatalogCheckpoint struct {
	ReceiptSHA256         string                           `json:"receipt_sha256"`
	ReceiptID             string                           `json:"receipt_id"`
	EnvironmentID         string                           `json:"environment_id"`
	Target                DatabaseIdentity                 `json:"target"`
	CatalogGeneration     int                              `json:"catalog_generation"`
	AdoptedAt             string                           `json:"adopted_at"`
	CatalogContractSHA256 string                           `json:"catalog_contract_sha256"`
	CatalogInstanceSHA256 string                           `json:"catalog_instance_sha256"`
	Objects               []AuthorityAdoptionReceiptObject `json:"objects"`
}

type AuthorityAdoptionReceipt struct {
	ContractVersion                string                              `json:"contract_version"`
	ReceiptID                      string                              `json:"receipt_id"`
	OperationID                    string                              `json:"operation_id"`
	ReleaseID                      string                              `json:"release_id"`
	EnvironmentID                  string                              `json:"environment_id"`
	Mode                           string                              `json:"mode"`
	PreviousCatalogGeneration      int                                 `json:"previous_catalog_generation"`
	CatalogGeneration              int                                 `json:"catalog_generation"`
	PreCatalogContractSHA256       string                              `json:"pre_catalog_contract_sha256"`
	PostCatalogContractSHA256      string                              `json:"post_catalog_contract_sha256"`
	PreviousAuthorityReceiptSHA256 *string                             `json:"previous_authority_receipt_sha256" harpyNull:"allowed"`
	UpgradeArtifactSHA256          *string                             `json:"upgrade_artifact_sha256" harpyNull:"allowed"`
	PreviousCatalogCheckpoint      *AuthorityCatalogCheckpoint         `json:"previous_catalog_checkpoint" harpyNull:"allowed"`
	Target                         DatabaseIdentity                    `json:"target"`
	OperationSHA256                string                              `json:"operation_sha256"`
	MigrationManifestSHA256        string                              `json:"migration_manifest_sha256"`
	ReleaseManifestSHA256          string                              `json:"release_manifest_sha256"`
	BootstrapAuthoritySHA256       string                              `json:"bootstrap_authority_manifest_sha256"`
	AdoptedAt                      string                              `json:"adopted_at"`
	Result                         string                              `json:"result"`
	ExecutorSessionUser            string                              `json:"executor_session_user"`
	ExecutorSuperuser              bool                                `json:"executor_superuser"`
	PreAdoptionMembershipCount     int                                 `json:"pre_adoption_membership_count"`
	PreAdoptionOwnedObjectCount    int                                 `json:"pre_adoption_owned_object_count"`
	AuthorityOwnerMembershipCount  int                                 `json:"authority_owner_membership_count"`
	QuiescenceEvidence             AuthorityQuiescenceEvidence         `json:"quiescence_evidence"`
	AuthorityOwnerRole             AuthorityOwnerRoleState             `json:"authority_owner_role"`
	CatalogInstanceSHA256          string                              `json:"catalog_instance_sha256"`
	Objects                        []AuthorityAdoptionReceiptObject    `json:"objects"`
	DependencyContractSHA256       string                              `json:"dependency_contract_sha256"`
	DependencyInstanceSHA256       string                              `json:"dependency_instance_sha256"`
	Dependencies                   []AuthorityCatalogDependencyReceipt `json:"dependencies"`
}

type AuthorityAdoptionSignature struct {
	ContractVersion string    `json:"contract_version"`
	ManifestSHA256  string    `json:"manifest_sha256"`
	Signature       Signature `json:"signature"`
}
