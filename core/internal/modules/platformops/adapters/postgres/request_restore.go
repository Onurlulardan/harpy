package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RequestRestore(
	ctx context.Context,
	record ports.RestoreRequestRecord,
) error {
	var safetyID, safetySHA, waiver, probe any
	if record.Safety.RecoverySetID != "" {
		safetyID = record.Safety.RecoverySetID
	}
	if record.Safety.RecoverySetSHA256 != "" {
		safetySHA = record.Safety.RecoverySetSHA256
	}
	if record.Safety.WaiverReason != "" {
		waiver = record.Safety.WaiverReason
	}
	if record.Safety.ProbeSHA256 != "" {
		probe = record.Safety.ProbeSHA256
	}
	err := store.withBound(ctx, transactionCommand, func(database boundDatabase) error {
		_, err := database.Execute(ctx, requestRestoreSQL,
			record.OperationID, record.RecoverySetID, record.EnvironmentID,
			string(record.Disposition), record.PlanSHA256, record.PlanBytes,
			safetyID, safetySHA, waiver, probe,
			record.Safety.DestructiveConfirmed, record.CreatedAt,
			record.RequestedByTenantID, record.RequestedByActorID,
			record.RequestedBySessionID)
		return store.mapDatabaseError(err)
	})
	if err != nil {
		return fmt.Errorf("request restore: %w", err)
	}
	return nil
}
