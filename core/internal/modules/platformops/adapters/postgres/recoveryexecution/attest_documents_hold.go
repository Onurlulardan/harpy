package recoveryexecution

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) AttestDocumentsHold(
	ctx context.Context,
	attestation ports.DocumentsHoldAttestation,
) (bool, error) {
	var changed bool
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		return database.QueryRow(ctx, attestDocumentsHoldSQL,
			attestation.RecoverySetID, attestation.AttemptID,
			attestation.LeaseToken, attestation.InventorySHA256,
			attestation.InventoryCount).Scan(&changed)
	})
	if err != nil {
		return false, fmt.Errorf("attest documents hold: %w", err)
	}
	return changed, nil
}
