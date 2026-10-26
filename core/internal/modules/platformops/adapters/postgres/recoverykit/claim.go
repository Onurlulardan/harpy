package recoverykit

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type Backend interface {
	RecoveryKitClaim(context.Context, string, string, string, time.Time) (ports.RecoveryKitClaimRecord, error)
}

type RecoveryKitOperations struct{ backend Backend }

func NewRecoveryKitOperations(backend Backend) *RecoveryKitOperations {
	return &RecoveryKitOperations{backend: backend}
}

func (operations *RecoveryKitOperations) ClaimRecoveryKit(
	ctx context.Context, operationID, actorUserID, sessionID string, occurredAt time.Time,
) (ports.RecoveryKitClaimRecord, error) {
	return operations.backend.RecoveryKitClaim(
		ctx, operationID, actorUserID, sessionID, occurredAt)
}
