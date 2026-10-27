package discountcatalog

import (
	"context"
	"fmt"
	"time"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
	"github.com/google/uuid"
)

func (storage *store) addOfferings(
	ctx context.Context,
	authorityID uuid.UUID,
	versionID uuid.UUID,
	policy discountdomain.Policy,
	occurredAt time.Time,
) error {
	for _, identifier := range policy.OfferingIDs() {
		if err := storage.execute(ctx, addOfferingSQL, authorityID, versionID,
			uuid.UUID(identifier), occurredAt); err != nil {
			return fmt.Errorf("add billing discount offering: %w", err)
		}
	}
	return nil
}
