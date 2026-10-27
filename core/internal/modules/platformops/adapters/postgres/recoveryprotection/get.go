package recoveryprotection

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type Backend interface {
	RecoveryProtectionLookup(context.Context, string) (ports.RecoveryProtectionRecord, error)
}

type RecoveryProtectionOperations struct{ backend Backend }

func NewRecoveryProtectionOperations(backend Backend) *RecoveryProtectionOperations {
	return &RecoveryProtectionOperations{backend: backend}
}

func (operations *RecoveryProtectionOperations) GetRecoveryProtection(
	ctx context.Context, environmentID string,
) (ports.RecoveryProtectionRecord, error) {
	return operations.backend.RecoveryProtectionLookup(ctx, environmentID)
}
