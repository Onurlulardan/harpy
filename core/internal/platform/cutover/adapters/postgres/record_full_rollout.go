package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/fullrollout"
	"github.com/google/uuid"
)

func (recorder *fullRolloutRecorder) Record(
	ctx context.Context,
	value fullrollout.PersistedCompletion,
	digest [sha256.Size]byte,
) ([sha256.Size]byte, error) {
	if recorder == nil || recorder.database == nil || len(value.Evidence.Steps) < 2 {
		return [sha256.Size]byte{}, fmt.Errorf("full rollout recorder is invalid")
	}
	evidence := value.Evidence
	tenantID := uuid.UUID(evidence.Steps[1].ScopeID)
	var persisted string
	err := recorder.database.QueryRow(ctx, recordFullRolloutCompletionSQL,
		value.ReceiptID, value.EnvironmentID, tenantID, encodeDigest(digest),
		encodeDigest(evidence.ShadowReadSHA256),
		encodeDigest(evidence.LogicalDeltaSHA256),
		encodeDigest(evidence.ReconcileSHA256),
		encodeDigest(evidence.ProductionSnapshotSHA256),
		encodeDigest(evidence.V1RemovalSHA256),
		encodeDigest(evidence.BackupRestoreSHA256),
		encodeDigest(evidence.FrontendContractSHA256),
		encodeDigest(evidence.RollbackWindowClosureSHA256), evidence.LastLegacyHitAt,
		evidence.RollbackWindowClosedAt, value.RecordedAt,
	).Scan(&persisted)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("persist full rollout completion: %w", err)
	}
	decoded, decodeErr := hex.DecodeString(persisted)
	if decodeErr != nil || len(decoded) != sha256.Size {
		return [sha256.Size]byte{}, fmt.Errorf("persisted full rollout digest is invalid")
	}
	var result [sha256.Size]byte
	copy(result[:], decoded)
	return result, nil
}

func encodeDigest(value [sha256.Size]byte) string { return hex.EncodeToString(value[:]) }
