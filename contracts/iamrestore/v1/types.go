package iamrestorev1

import "encoding/json"

const (
	SecurityEvidenceSchema = "iam-restore-security-evidence.v1"
	CustodyManifestSchema  = "iam-key-custody-manifest.v1"
	MaximumDocumentBytes   = 512 * 1024
)

type SecurityEvidence struct {
	Schema                    string          `json:"schema"`
	EnvironmentID             string          `json:"environment_id"`
	RestoreOperationID        string          `json:"restore_operation_id"`
	RecoverySetID             string          `json:"recovery_set_id"`
	TenantID                  string          `json:"tenant_id"`
	OwnerUserID               string          `json:"owner_user_id"`
	OwnerBindingID            string          `json:"owner_binding_id"`
	OwnerBindingGeneration    int64           `json:"owner_binding_generation"`
	AnchorID                  string          `json:"anchor_id"`
	NewDatabaseSystemID       string          `json:"new_database_system_id"`
	RestoreReceiptSHA256      string          `json:"restore_receipt_sha256"`
	RecoveryTrustSHA256       string          `json:"recovery_trust_bundle_sha256"`
	AckTrustSHA256            string          `json:"ack_trust_bundle_sha256"`
	CustodyManifestSHA256     string          `json:"iam_custody_manifest_sha256"`
	ProtectedHeadSHA256       string          `json:"protected_head_sha256"`
	ProtectedAckSHA256        string          `json:"protected_ack_sha256"`
	ProtectedCheckpointSHA256 string          `json:"protected_checkpoint_sha256"`
	ProtectedAck              json.RawMessage `json:"protected_ack"`
	ProtectedCheckpoint       json.RawMessage `json:"protected_checkpoint"`
}

type CustodyManifest struct {
	Schema        string           `json:"schema"`
	EnvironmentID string           `json:"environment_id"`
	Provider      string           `json:"provider"`
	KeyringHandle string           `json:"keyring_handle"`
	Rotations     []RotationResult `json:"rotations"`
}

type RotationResult struct {
	OperationID        string   `json:"operation_id"`
	Status             string   `json:"status"`
	Generation         int64    `json:"generation"`
	SecurityEpoch      string   `json:"security_epoch"`
	ResultSHA256       string   `json:"result_sha256"`
	PublicKIDs         []string `json:"public_kids"`
	RevokedGenerations []int64  `json:"revoked_generations"`
}
