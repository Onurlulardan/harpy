package recoveryexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (service *Service) AttestDocuments(
	ctx context.Context,
	attestation ports.DocumentsHoldAttestation,
) error {
	if service == nil || ctx == nil || ctx.Err() != nil ||
		!validIdentity(attestation.RecoveryExecutionIdentity) ||
		!validDigest(attestation.InventorySHA256) || attestation.InventoryCount < 0 {
		return ErrInvalid
	}
	changed, err := service.store.AttestDocumentsHold(ctx, attestation)
	if err != nil {
		return ErrUnavailable
	}
	if !changed {
		return ErrConflict
	}
	return nil
}
