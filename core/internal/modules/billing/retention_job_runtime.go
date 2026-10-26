package billing

import (
	"context"
	"fmt"

	retentionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/retention"
	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/retention"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type RetentionJobRuntime struct {
	binding platformasync.DatabaseJobBinding
}

func NewRetentionJobRuntime(
	nextID retentionapp.IDSource,
) (*RetentionJobRuntime, error) {
	if nextID == nil {
		return nil, retentionapp.ErrConfiguration
	}
	binding, err := platformasync.NewDatabaseJobBinding(
		"billing",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			store, storeErr := retentionpostgres.NewStore(retentionpostgres.DatabaseAccess{
				QueryRow: func(
					ctx context.Context, statement string, arguments ...any,
				) retentionpostgres.Row {
					return database.QueryRow(ctx, statement, arguments...)
				},
			})
			commands, commandErr := retentionapp.NewCommands(store, nextID)
			if storeErr != nil {
				commandErr = storeErr
			}
			return &retentionDatabaseJobCommand{commands: commands, buildErr: commandErr}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("build billing retention job binding: %w", err)
	}
	return &RetentionJobRuntime{binding: binding}, nil
}

func (runtime *RetentionJobRuntime) Binding() platformasync.DatabaseJobBinding {
	if runtime == nil {
		return platformasync.DatabaseJobBinding{}
	}
	return runtime.binding
}
