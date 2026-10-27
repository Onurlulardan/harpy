package recoveryset

import "time"

type DeploymentMode string
type Portability string
type TargetDisposition string

const (
	ModeOnPremise DeploymentMode = "onpremise"
	ModeSaaS      DeploymentMode = "saas"

	SameEnvironment Portability = "same_environment"
	Portable        Portability = "portable"

	TargetExistingReadable TargetDisposition = "existing_readable"
	TargetEmpty            TargetDisposition = "empty"
	TargetUnreadable       TargetDisposition = "unreadable"
)

type EnvironmentIdentity struct {
	EnvironmentID          string         `json:"environment_id"`
	DeploymentMode         DeploymentMode `json:"deployment_mode"`
	PrimaryBindingID       string         `json:"primary_binding_id"`
	TenantSetSHA256        string         `json:"tenant_set_sha256"`
	OwnerBindingID         string         `json:"owner_binding_id"`
	OwnerBindingGeneration int64          `json:"owner_binding_generation"`
	DatabaseLineageID      string         `json:"database_lineage_id"`
}

type ComponentIdentity struct {
	ObjectKey        string `json:"object_key"`
	ObjectVersionID  string `json:"object_version_id"`
	CiphertextSHA256 string `json:"ciphertext_sha256"`
	CiphertextBytes  int64  `json:"ciphertext_bytes"`
}

type RecoveryComponents struct {
	DatabaseArchive    ComponentIdentity `json:"database_archive"`
	DocumentsInventory ComponentIdentity `json:"documents_inventory"`
	KeyCustody         ComponentIdentity `json:"key_custody"`
	OwnerAnchor        ComponentIdentity `json:"owner_anchor"`
	ProtectedAuthority ComponentIdentity `json:"protected_authority"`
}

type ManifestPayload struct {
	Schema                    string              `json:"schema"`
	Portability               Portability         `json:"portability"`
	Identity                  EnvironmentIdentity `json:"identity"`
	DatabaseSystemID          string              `json:"database_system_id"`
	Components                RecoveryComponents  `json:"components"`
	DatabaseArchiveSHA256     string              `json:"database_archive_sha256"`
	DocumentsMerkleSHA256     string              `json:"documents_merkle_sha256"`
	KeyCustodySHA256          string              `json:"key_custody_sha256"`
	OwnerAnchorSHA256         string              `json:"owner_anchor_sha256"`
	ProtectedHeadAckSHA256    string              `json:"protected_head_ack_sha256"`
	RecoveryTrustBundleSHA256 string              `json:"recovery_trust_bundle_sha256"`
	RecoverySigningGeneration int64               `json:"recovery_signing_generation"`
	CreatedAt                 time.Time           `json:"created_at"`
}

type SafetyEvidence struct {
	RecoverySetID        string
	RecoverySetSHA256    string
	WaiverReason         string
	ProbeSHA256          string
	DestructiveConfirmed bool
}

type RestoreRequest struct {
	RecoverySetID     string
	Manifest          ManifestPayload
	Target            EnvironmentIdentity
	Disposition       TargetDisposition
	Safety            SafetyEvidence
	SeedRequested     bool
	OwnerChange       bool
	TenantReplacement bool
}
