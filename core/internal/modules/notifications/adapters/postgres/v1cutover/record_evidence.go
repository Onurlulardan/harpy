package v1cutover

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
	"github.com/google/uuid"
)

var evidenceNamespace = uuid.MustParse("a989668a-c86a-485a-94c5-0fa0c65bfc7a")

func (store *Store) RecordEvidence(ctx context.Context,
	inventory app.Inventory,
) (app.Evidence, error) {
	if !store.usable(ctx) || inventory.Validate() != nil {
		return app.Evidence{}, app.ErrInvalid
	}
	identity := append([]byte(inventory.RunID.String()+"\x00"),
		inventory.SourceEvidenceDigest[:]...)
	evidenceID := uuid.NewSHA1(evidenceNamespace, identity)
	var storedID uuid.UUID
	var expiresAt = inventory.ObservedAt
	err := store.queryRow(ctx, recordEvidenceSQL, evidenceID, inventory.RunID,
		inventory.SourceQueueDepth, inventory.FutureDeliveryCount,
		inventory.RunningDeliveryCount, inventory.UnresolvedProviderReceiptCount,
		inventory.LegacyConnectionCount, inventory.SourceEvidenceDigest[:],
		inventory.ObservedAt.UTC()).Scan(&storedID, &expiresAt)
	if err != nil || storedID == uuid.Nil {
		return app.Evidence{}, app.ErrConflict
	}
	return app.Evidence{ID: storedID, Inventory: inventory,
		ExpiresAt: expiresAt.UTC()}, nil
}
