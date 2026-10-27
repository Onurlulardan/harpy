package billing

import (
	"context"
	"fmt"

	providerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/provider"
	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type providerFinalizeDatabaseJobCommand struct {
	database platformtx.JobBoundExecutor
	nextID   providerapp.CommandIDSource
	clock    providerapp.CommandClock
}

func (command *providerFinalizeDatabaseJobCommand) Execute(
	ctx context.Context,
	claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.nextID == nil ||
		command.clock == nil || ctx == nil || claim == nil {
		return platformasync.DatabaseJobOutcome{}, providerapp.ErrConfiguration
	}
	request, err := providerFinalizeRequest(claim)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	store, err := providerpostgres.NewBoundFinalizerStore(
		request.TenantID, command.database,
	)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	finalizer, err := providerapp.NewCommandFinalizer(
		store, command.nextID, command.clock,
	)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	result, err := finalizer.FinalizeCommand(ctx, request)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"execute billing provider finalizer: %w", err,
		)
	}
	return providerFinalizationOutcome(claim, result, command.nextID)
}

var _ platformasync.DatabaseJobCommand = (*providerFinalizeDatabaseJobCommand)(nil)
