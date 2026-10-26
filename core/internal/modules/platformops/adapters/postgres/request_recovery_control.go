package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func (store *Store) RequestRecoveryControl(
	ctx context.Context,
	request ports.RecoveryControlRequestRecord,
) (ports.RecoveryControlOperationRecord, error) {
	record := ports.RecoveryControlOperationRecord{
		EnvironmentID: request.EnvironmentID, Kind: request.Kind,
		CreatedAt: request.OccurredAt,
	}
	err := store.withBound(ctx, transactionCommand,
		func(database boundDatabase) error {
			err := database.QueryRow(ctx, requestRecoveryControlSQL,
				request.OperationID, request.EnvironmentID, request.ActorUserID,
				request.IdempotencyKey, request.RequestSHA256, request.Kind,
				nullableText(request.ClaimAlgorithm), nullableBytes(request.ClaimPublicKey),
				request.OccurredAt,
			).Scan(&record.OperationID, &record.State, &record.KitClaimState)
			return store.mapDatabaseError(err)
		})
	if err != nil {
		return ports.RecoveryControlOperationRecord{},
			fmt.Errorf("request recovery control: %w", err)
	}
	return record, nil
}

func nullableText(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}
