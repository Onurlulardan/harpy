package billing

import (
	"context"
	"fmt"

	subscriptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/subscription"
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type subscriptionJobCommand struct {
	database platformtx.JobBoundExecutor
	nextID   subscriptionapp.IDGenerator
	clock    subscriptionapp.Clock
}

func (command *subscriptionJobCommand) Execute(
	ctx context.Context,
	claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.nextID == nil ||
		command.clock == nil || ctx == nil || claim == nil {
		return platformasync.DatabaseJobOutcome{}, subscriptionapp.ErrConfiguration
	}
	tenantID, tenantScoped := claim.TenantID()
	if !tenantScoped || claim.LeaseGeneration() < 1 {
		return platformasync.DatabaseJobOutcome{}, subscriptionapp.ErrRequest
	}
	intent, err := subscriptionBoundaryIntent(tenantID, claim)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	effects := newDeferredSubscriptionEffects(tenantID)
	store, err := subscriptionpostgres.NewBoundJobStore(
		tenantID, command.database, effects,
	)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	commands, err := subscriptionapp.NewCommands(store, command.nextID, command.clock)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	_, err = commands.ExecuteDueBoundary(ctx, subscriptionapp.ExecuteDueBoundaryRequest{
		SubscriptionID: intent.SubscriptionID, JobID: claim.JobID(),
		EffectID: intent.EffectID, ExpectedRevision: intent.ExpectedRevision,
		LeaseGeneration: claim.LeaseGeneration(), Boundary: intent.Boundary,
		DueAt: intent.RunAt,
	})
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"execute billing subscription boundary job: %w", err,
		)
	}
	return effects.outcome()
}

var _ platformasync.DatabaseJobCommand = (*subscriptionJobCommand)(nil)
