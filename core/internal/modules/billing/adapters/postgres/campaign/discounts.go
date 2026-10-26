package campaign

import (
	"context"
	"fmt"
	"time"

	campaigndomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/campaign"
	"github.com/google/uuid"
)

func (storage *store) addDiscounts(
	ctx context.Context,
	authorityID uuid.UUID,
	versionID uuid.UUID,
	publication campaigndomain.Publication,
	occurredAt time.Time,
) error {
	for _, identifier := range publication.DiscountVersionIDs() {
		if err := storage.execute(ctx, addDiscountSQL, authorityID, versionID,
			uuid.UUID(identifier), occurredAt); err != nil {
			return fmt.Errorf("add billing campaign discount: %w", err)
		}
	}
	return nil
}
