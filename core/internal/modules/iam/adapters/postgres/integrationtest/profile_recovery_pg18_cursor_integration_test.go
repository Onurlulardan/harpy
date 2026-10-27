//go:build integration

package integrationtest

import (
	"context"

	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func drainRecoveryInventory(
	ctx context.Context,
	reader *profilebackfillpostgres.RecoveryReader,
	scope platformtx.TenantScope,
	parentID uuid.UUID,
	evidence profilebackfillapp.Evidence,
) ([]profilebackfillpostgres.RecoveryInventoryRow, error) {
	cursor, err := reader.OpenInventory(ctx, scope, parentID, evidence)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()
	var result []profilebackfillpostgres.RecoveryInventoryRow
	for {
		batch, fetchErr := cursor.Fetch(ctx)
		if fetchErr != nil {
			return nil, fetchErr
		}
		result = append(result, batch...)
		if len(batch) < profilebackfillapp.RecoveryInventoryReadBatchSize {
			if closeErr := cursor.Close(ctx); closeErr != nil {
				return nil, closeErr
			}
			return result, nil
		}
	}
}
