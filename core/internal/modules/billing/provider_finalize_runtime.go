package billing

import (
	"fmt"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func newProviderFinalizeJobBinding() (
	platformasync.DatabaseJobBinding,
	error,
) {
	binding, err := platformasync.NewDatabaseJobBinding(
		"billing",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			return &providerFinalizeDatabaseJobCommand{
				database: database, nextID: uuid.NewV7,
				clock: func() time.Time {
					return time.Now().UTC().Truncate(time.Microsecond)
				},
			}
		},
	)
	if err != nil {
		return platformasync.DatabaseJobBinding{},
			fmt.Errorf("build billing provider finalizer binding: %w", err)
	}
	if binding.Owner() != "billing" {
		return platformasync.DatabaseJobBinding{}, providerapp.ErrConfiguration
	}
	return binding, nil
}
