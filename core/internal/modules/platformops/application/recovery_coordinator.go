package application

import "github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"

type RecoveryCoordinator struct {
	store ports.RecoveryStore
	clock ports.Clock
}

func NewRecoveryCoordinator(store ports.RecoveryStore, clock ports.Clock) *RecoveryCoordinator {
	if store == nil || clock == nil {
		return nil
	}
	return &RecoveryCoordinator{store: store, clock: clock}
}
