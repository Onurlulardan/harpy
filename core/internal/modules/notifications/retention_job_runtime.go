package notifications

import (
	"context"
	"fmt"

	retentionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/postgres"
	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type RetentionJobRuntime struct {
	binding platformasync.DatabaseJobBinding
}

func NewRetentionJobRuntime(nextID retentionapp.IDSource,
	clock retentionapp.Clock,
) (*RetentionJobRuntime, error) {
	if nextID == nil || clock == nil {
		return nil, retentionapp.ErrConfiguration
	}
	binding, err := platformasync.NewDatabaseJobBinding("notifications",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			store, storeErr := retentionpostgres.NewRetentionStore(retentionpostgres.RetentionDatabaseAccess{
				QueryRow: func(ctx context.Context, statement string,
					arguments ...any,
				) retentionpostgres.RetentionRow {
					return database.QueryRow(ctx, statement, arguments...)
				},
				RowNotFound: retentionpostgres.IsRowNotFound,
			})
			commands, commandErr := retentionapp.NewCommands(store, nextID, clock)
			if storeErr != nil {
				commandErr = storeErr
			}
			return &retentionDatabaseJobCommand{commands: commands, buildErr: commandErr}
		})
	if err != nil {
		return nil, fmt.Errorf("build notification retention binding: %w", err)
	}
	return &RetentionJobRuntime{binding: binding}, nil
}

func (runtime *RetentionJobRuntime) Binding() platformasync.DatabaseJobBinding {
	if runtime == nil {
		return platformasync.DatabaseJobBinding{}
	}
	return runtime.binding
}
