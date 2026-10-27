package recoveryset

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type Backend interface {
	RecoverySetListing(context.Context, string) ([]ports.RecoverySetRecord, error)
}

type RecoverySetOperations struct{ backend Backend }

func NewRecoverySetOperations(backend Backend) *RecoverySetOperations {
	return &RecoverySetOperations{backend: backend}
}

func (operations *RecoverySetOperations) ListRecoverySets(
	ctx context.Context, environmentID string,
) ([]ports.RecoverySetRecord, error) {
	return operations.backend.RecoverySetListing(ctx, environmentID)
}
