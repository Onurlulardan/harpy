package recoverykit

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	RecoveryKitClaim(context.Context, string, string, string) (platformopsapi.RecoveryKitClaim, error)
}

type RecoveryKitOperations struct{ backend Backend }

func NewRecoveryKitOperations(backend Backend) *RecoveryKitOperations {
	return &RecoveryKitOperations{backend: backend}
}

func (operations *RecoveryKitOperations) ClaimKit(
	ctx context.Context, operationID, actorUserID, sessionID string,
) (platformopsapi.RecoveryKitClaim, error) {
	return operations.backend.RecoveryKitClaim(ctx, operationID, actorUserID, sessionID)
}
