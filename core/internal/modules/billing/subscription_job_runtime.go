package billing

import (
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type SubscriptionJobRuntime struct {
	binding platformasync.DatabaseJobBinding
}

func NewSubscriptionJobRuntime(
	nextID subscriptionapp.IDGenerator,
	clock subscriptionapp.Clock,
) (*SubscriptionJobRuntime, error) {
	if nextID == nil || clock == nil {
		return nil, subscriptionapp.ErrConfiguration
	}
	binding, err := platformasync.NewDatabaseJobBinding(
		"billing",
		func(database platformtx.JobBoundExecutor) platformasync.DatabaseJobCommand {
			return &subscriptionJobCommand{
				database: database, nextID: nextID, clock: clock,
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("build billing subscription job binding: %w", err)
	}
	return &SubscriptionJobRuntime{binding: binding}, nil
}

func (runtime *SubscriptionJobRuntime) Binding() platformasync.DatabaseJobBinding {
	if runtime == nil {
		return platformasync.DatabaseJobBinding{}
	}
	return runtime.binding
}
