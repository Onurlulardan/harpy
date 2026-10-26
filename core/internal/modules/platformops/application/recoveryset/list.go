package recoveryset

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	RecoverySetListing(context.Context, string) ([]platformopsapi.RecoverySet, error)
}

type RecoverySetOperations struct{ backend Backend }

func NewRecoverySetOperations(backend Backend) *RecoverySetOperations {
	return &RecoverySetOperations{backend: backend}
}

func (operations *RecoverySetOperations) ListRecoverySets(
	ctx context.Context, environmentID string,
) ([]platformopsapi.RecoverySet, error) {
	return operations.backend.RecoverySetListing(ctx, environmentID)
}
