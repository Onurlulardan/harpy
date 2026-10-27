package recoveryprotection

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	RecoveryProtectionLookup(context.Context, string) (platformopsapi.RecoveryProtection, error)
}

type RecoveryProtectionOperations struct{ backend Backend }

func NewRecoveryProtectionOperations(backend Backend) *RecoveryProtectionOperations {
	return &RecoveryProtectionOperations{backend: backend}
}

func (operations *RecoveryProtectionOperations) GetProtection(
	ctx context.Context, environmentID string,
) (platformopsapi.RecoveryProtection, error) {
	return operations.backend.RecoveryProtectionLookup(ctx, environmentID)
}
