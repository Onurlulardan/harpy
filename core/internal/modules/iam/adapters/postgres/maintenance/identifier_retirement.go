package maintenance

import (
	"context"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func (store *store) executeIdentifierRetire(
	ctx context.Context, claim maintenanceapp.Claim, nextID uuid.UUID,
) (maintenanceapp.Outcome, error) {
	return store.executeSimple(ctx, claim, nextID, executeIdentifierRetireSQL)
}
