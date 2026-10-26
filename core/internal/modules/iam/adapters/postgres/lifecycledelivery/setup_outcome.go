package lifecycledelivery

import (
	"context"

	"github.com/google/uuid"
)

func (store *store) recordSetupOutcome(ctx context.Context,
	tenantID, artifactID uuid.UUID, outcome string,
) error {
	if store == nil || store.database == nil || ctx == nil ||
		tenantID == uuid.Nil || artifactID == uuid.Nil {
		return nil
	}
	var replay bool
	return store.database.QueryRow(ctx, recordSetupOutcomeSQL,
		tenantID, artifactID, outcome).Scan(&replay)
}
