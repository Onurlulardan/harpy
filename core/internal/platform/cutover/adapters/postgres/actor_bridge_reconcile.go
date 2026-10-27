package postgres

import (
	"context"
	"database/sql"

	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
)

func (database *ActorBridgeDatabase) ReconcileHits(
	ctx context.Context,
	limit int,
) (actorbridge.Reconciliation, error) {
	if database == nil || actorBridgeQueryerIsNil(database.queryer) || ctx == nil ||
		ctx.Err() != nil || limit < 1 || limit > 10000 {
		return actorbridge.Reconciliation{}, actorbridge.NewFault(actorbridge.FaultInvalid)
	}
	var count int64
	var latest sql.NullTime
	err := database.queryer.QueryRow(ctx, reconcileIAMV1ActorAssertionHitsSQL,
		limit).Scan(&count, &latest)
	if err != nil {
		return actorbridge.Reconciliation{}, classifyActorBridgeError(err)
	}
	if !latest.Valid {
		return actorbridge.NewReconciliation(count, nil)
	}
	return actorbridge.NewReconciliation(count, &latest.Time)
}
