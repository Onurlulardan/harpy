package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type RecoveryService struct {
	sets       ports.RecoverySetStore
	restores   ports.RestoreStore
	signing    ports.RecoveryStore
	identities ports.EnvironmentIdentityReader
	planSigner ports.RestorePlanSigner
	clock      ports.Clock
}

func NewRecoveryServiceWithSigner(
	sets ports.RecoverySetStore,
	restores ports.RestoreStore,
	signing ports.RecoveryStore,
	identities ports.EnvironmentIdentityReader,
	planSigner ports.RestorePlanSigner,
	clock ports.Clock,
) *RecoveryService {
	service := NewRecoveryService(sets, restores, signing, identities, clock)
	if service == nil || planSigner == nil {
		return nil
	}
	service.planSigner = planSigner
	return service
}

func NewRecoveryService(
	sets ports.RecoverySetStore,
	restores ports.RestoreStore,
	signing ports.RecoveryStore,
	identities ports.EnvironmentIdentityReader,
	clock ports.Clock,
) *RecoveryService {
	if sets == nil || restores == nil || signing == nil || identities == nil || clock == nil {
		return nil
	}
	return &RecoveryService{sets: sets, restores: restores,
		signing: signing, identities: identities, clock: clock}
}
