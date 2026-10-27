package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
	"github.com/pressly/goose/v3"
)

func (runner *Runner) bootstrapRunLifecycle(
	ctx context.Context,
	connection *sql.Conn,
	provider *goose.Provider,
) error {
	ready, err := postgresstore.MigrationRunInfrastructureReady(ctx, connection)
	if err != nil {
		return err
	}
	if !ready {
		if _, err := provider.UpTo(ctx, runLifecycleReadyVersion); err != nil {
			return fmt.Errorf("apply migration run infrastructure: %w", err)
		}
		if err := requireLifecycleReady(ctx, connection); err != nil {
			return err
		}
		history, _, err := readHistory(ctx, connection)
		if err != nil {
			return err
		}
		_, pending, err := classifyHistory(history, runner.selected, runner.historyAliases)
		if err != nil {
			return err
		}
		for _, migration := range pending {
			if migration.Version <= runLifecycleReadyVersion {
				return fmt.Errorf("migration run infrastructure history is incomplete")
			}
		}
	}
	// Platform cutover migrations consume the verified identity before finalization.
	if err := runner.persistEnvironmentIdentity(ctx, connection); err != nil {
		return fmt.Errorf("persist bootstrap environment identity: %w", err)
	}
	return nil
}
