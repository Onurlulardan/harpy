package artifact

type State struct {
	Schema                         string         `json:"schema"`
	EnvironmentID                  string         `json:"environment_id"`
	DeploymentMode                 DeploymentMode `json:"deployment_mode"`
	RegistryBindingID              string         `json:"registry_binding_id"`
	ScopeKind                      ScopeKind      `json:"scope_kind"`
	ScopeID                        string         `json:"scope_id"`
	ProposalSignerGeneration       int64          `json:"proposal_signer_generation"`
	ProposalSignerKID              string         `json:"proposal_signer_kid"`
	OwnerBindingID                 *string        `json:"owner_binding_id"`
	OwnerBindingGeneration         *int64         `json:"owner_binding_generation"`
	PrimaryAdminBindingSetSHA256   *string        `json:"primary_protected_admin_binding_set_sha256"`
	RecoveryAnchorSetSHA256        *string        `json:"recovery_anchor_set_sha256"`
	RecoveryTrustBundleSHA256      *string        `json:"recovery_trust_bundle_sha256"`
	TenantID                       *string        `json:"tenant_id"`
	ProtectedAdminBindingSetSHA256 *string        `json:"protected_admin_binding_set_sha256"`
}

type Proposal struct {
	Schema                    string                  `json:"schema"`
	Purpose                   string                  `json:"purpose"`
	OperationID               string                  `json:"operation_id"`
	EnvironmentID             string                  `json:"environment_id"`
	DeploymentMode            DeploymentMode          `json:"deployment_mode"`
	ScopeKind                 ScopeKind               `json:"scope_kind"`
	ScopeID                   string                  `json:"scope_id"`
	RegistryBindingID         string                  `json:"registry_binding_id"`
	Generation                int64                   `json:"generation"`
	PreviousHeadSHA256        *string                 `json:"previous_head_sha256"`
	StateSHA256               string                  `json:"state_sha256"`
	RecoveryTrustBundleSHA256 *string                 `json:"recovery_trust_bundle_sha256"`
	CompensatesAckSHA256      *string                 `json:"compensates_ack_sha256"`
	IssuedAt                  string                  `json:"issued_at"`
	ExpiresAt                 string                  `json:"expires_at"`
	SignerKID                 string                  `json:"signer_kid"`
	SignerRotation            *SignerRotationEvidence `json:"signer_rotation,omitempty"`
	Signature                 string                  `json:"signature"`
}

type SignerRotationEvidence struct {
	PreviousSignerGeneration int64   `json:"previous_signer_generation"`
	NewSignerKID             string  `json:"new_signer_kid"`
	NewSignerPublicKeySHA256 string  `json:"new_signer_public_key_sha256"`
	RecoveryTrustBundleBytes *string `json:"recovery_trust_bundle_bytes"`
}

type Ack struct {
	Schema             string       `json:"schema"`
	RegistryKind       RegistryKind `json:"registry_kind"`
	RegistryBindingID  string       `json:"registry_binding_id"`
	ScopeKind          ScopeKind    `json:"scope_kind"`
	ScopeID            string       `json:"scope_id"`
	RegistrySequence   int64        `json:"registry_sequence"`
	OperationID        string       `json:"operation_id"`
	ProposalSHA256     string       `json:"proposal_sha256"`
	Generation         int64        `json:"generation"`
	PreviousHeadSHA256 *string      `json:"previous_head_sha256"`
	StateSHA256        string       `json:"state_sha256"`
	IssuedAt           string       `json:"issued_at"`
	RegistryKID        string       `json:"registry_kid"`
	Signature          string       `json:"signature"`
}
