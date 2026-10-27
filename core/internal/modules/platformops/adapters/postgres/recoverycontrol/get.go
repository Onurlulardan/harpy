package recoverycontrol

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type Backend interface {
	RecoveryControlLookup(context.Context, string) (ports.RecoveryControlOperationRecord, error)
}

type RecoveryControlOperations struct{ backend Backend }

func NewRecoveryControlOperations(backend Backend) *RecoveryControlOperations {
	return &RecoveryControlOperations{backend: backend}
}

func (operations *RecoveryControlOperations) GetRecoveryControlOperation(
	ctx context.Context, operationID string,
) (ports.RecoveryControlOperationRecord, error) {
	return operations.backend.RecoveryControlLookup(ctx, operationID)
}
