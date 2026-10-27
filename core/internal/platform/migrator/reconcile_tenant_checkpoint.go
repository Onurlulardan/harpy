package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func reconcileTenantCheckpoint(
	ctx context.Context,
	transaction *sql.Tx,
	checkpoint postgresstore.TenantCheckpoint,
	runID string,
) (postgresstore.TenantCheckpoint, error) {
	switch checkpoint.State {
	case "pending":
		if checkpoint.RunID != runID {
			return checkpoint, fmt.Errorf("pending tenant checkpoint belongs to another run")
		}
	case "running":
		if checkpoint.RunID == runID {
			return checkpoint, fmt.Errorf("tenant checkpoint is already running for this run")
		}
		identity := checkpoint.TenantCheckpointIdentity
		if err := postgresstore.CompleteTenantCheckpoint(
			ctx, transaction, identity, "failed", "migration_interrupted",
		); err != nil {
			return checkpoint, err
		}
		checkpoint.State = "failed"
	case "failed":
	case "fenced":
		return checkpoint, fmt.Errorf("tenant checkpoint is fenced")
	default:
		return checkpoint, fmt.Errorf("tenant checkpoint state is invalid")
	}
	if checkpoint.State == "failed" {
		if checkpoint.RunID == runID {
			return checkpoint, fmt.Errorf("tenant retry requires a new migration run")
		}
		checkpoint.Attempt++
	}
	return checkpoint, nil
}
