package analytics

import (
	"context"
	"fmt"

	retentionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/analytics/adapters/postgres/retention"
	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/retention"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func newAnalyticsRetentionBinding() (platformasync.DatabaseJobBinding, error) {
	binding, err := platformasync.NewDatabaseJobBinding("analytics",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			store, storeErr := retentionpostgres.NewStore(retentionpostgres.DatabaseAccess{
				QueryRow: func(
					ctx context.Context, statement string, arguments ...any,
				) retentionpostgres.Row {
					return database.QueryRow(ctx, statement, arguments...)
				},
			})
			commands, commandErr := retentionapp.NewCommands(store, uuid.NewV7)
			if storeErr != nil {
				commandErr = storeErr
			}
			return &analyticsRetentionCommand{commands: commands, buildErr: commandErr}
		})
	if err != nil {
		return platformasync.DatabaseJobBinding{}, fmt.Errorf(
			"build Analytics retention binding: %w", err)
	}
	return binding, nil
}
