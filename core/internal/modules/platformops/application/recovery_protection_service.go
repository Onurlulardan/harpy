package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type RecoveryProtectionService struct {
	store ports.RecoveryProtectionStore
	clock ports.Clock
}

func NewRecoveryProtectionService(
	store ports.RecoveryProtectionStore,
	clock ports.Clock,
) *RecoveryProtectionService {
	if store == nil || clock == nil {
		return nil
	}
	return &RecoveryProtectionService{store: store, clock: clock}
}
