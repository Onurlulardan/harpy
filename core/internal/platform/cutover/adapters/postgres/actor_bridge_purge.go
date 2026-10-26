package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
)

func (database *ActorBridgeDatabase) Purge(ctx context.Context, limit int) (int64, error) {
	if database == nil || actorBridgeQueryerIsNil(database.queryer) || ctx == nil ||
		ctx.Err() != nil || limit < 1 || limit > 10000 {
		return 0, actorbridge.NewFault(actorbridge.FaultInvalid)
	}
	var deleted int64
	if err := database.queryer.QueryRow(ctx, purgeIAMV1ActorAssertionsSQL,
		limit).Scan(&deleted); err != nil {
		return 0, classifyActorBridgeError(err)
	}
	if deleted < 0 || deleted > int64(limit) {
		return 0, actorbridge.NewFault(actorbridge.FaultUnavailable)
	}
	return deleted, nil
}
