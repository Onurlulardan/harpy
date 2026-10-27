package iam

import (
	"encoding/hex"
	"encoding/json"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type RestoredIdentityReceiptInput struct {
	RestoreOperationID   UUID
	EnvironmentID        UUID
	RecoverySetID        UUID
	TenantID             UUID
	OwnerUserID          UUID
	AnchorID             UUID
	NewDatabaseSystemID  UUID
	SecurityEvidence     RestoredSecurityEvidence
	RestoreReceiptSHA256 [32]byte
	SecurityEpochSHA256  [32]byte
	Custody              RestoreCustodyResult
	OccurredAt           time.Time
}

type restoredIdentityReceipt struct {
	Schema                    string   `json:"schema"`
	RestoreOperationID        string   `json:"restore_operation_id"`
	EnvironmentID             string   `json:"environment_id"`
	RecoverySetID             string   `json:"recovery_set_id"`
	TenantID                  string   `json:"tenant_id"`
	OwnerUserID               string   `json:"owner_user_id"`
	AnchorID                  string   `json:"anchor_id"`
	NewDatabaseSystemID       string   `json:"new_database_system_id"`
	SecurityEpochSHA256       string   `json:"security_epoch_sha256"`
	RestoreReceiptSHA256      string   `json:"restore_receipt_sha256"`
	ProtectedHeadSHA256       string   `json:"protected_head_sha256"`
	ProtectedAckSHA256        string   `json:"protected_ack_sha256"`
	ProtectedCheckpointSHA256 string   `json:"protected_checkpoint_sha256"`
	CustodyGeneration         int64    `json:"custody_generation"`
	CustodyResultSHA256       string   `json:"custody_result_sha256"`
	CustodyPublicKIDs         []string `json:"custody_public_kids"`
	RevokedGenerations        []int64  `json:"revoked_generations"`
	OccurredAt                string   `json:"occurred_at"`
}

func BuildRestoredIdentityReceipt(input RestoredIdentityReceiptInput) ([]byte, error) {
	security := input.SecurityEvidence
	rotation := input.Custody
	value := restoredIdentityReceipt{
		Schema:             "iam-restored-identity-receipt.v1",
		RestoreOperationID: input.RestoreOperationID.String(),
		EnvironmentID:      input.EnvironmentID.String(), RecoverySetID: input.RecoverySetID.String(),
		TenantID: input.TenantID.String(), OwnerUserID: input.OwnerUserID.String(),
		AnchorID: input.AnchorID.String(), NewDatabaseSystemID: input.NewDatabaseSystemID.String(),
		SecurityEpochSHA256:       hex.EncodeToString(input.SecurityEpochSHA256[:]),
		RestoreReceiptSHA256:      hex.EncodeToString(input.RestoreReceiptSHA256[:]),
		ProtectedHeadSHA256:       security.ProtectedHeadSHA256,
		ProtectedAckSHA256:        security.ProtectedAckSHA256,
		ProtectedCheckpointSHA256: security.ProtectedCheckpointSHA256,
		CustodyGeneration:         rotation.Generation, CustodyResultSHA256: rotation.ResultSHA256,
		CustodyPublicKIDs: rotation.PublicKIDs, RevokedGenerations: rotation.RevokedGenerations,
		OccurredAt: input.OccurredAt.Format(time.RFC3339Nano),
	}
	document, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(document, 64*1024)
}
