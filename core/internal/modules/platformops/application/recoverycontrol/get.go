package recoverycontrol

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	RecoveryControlLookup(context.Context, string) (platformopsapi.RecoveryControlOperation, error)
}

type RecoveryControlOperations struct{ backend Backend }

func NewRecoveryControlOperations(backend Backend) *RecoveryControlOperations {
	return &RecoveryControlOperations{backend: backend}
}

func (operations *RecoveryControlOperations) GetControlOperation(
	ctx context.Context, operationID string,
) (platformopsapi.RecoveryControlOperation, error) {
	return operations.backend.RecoveryControlLookup(ctx, operationID)
}
