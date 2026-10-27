package transaction

import (
	"context"
	"slices"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func prepareGlobalScope(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	request GlobalScopeRequest,
) (GlobalScope, error) {
	environmentID, err := database.ResolveGlobalEnvironmentID(ctx)
	if err != nil {
		return GlobalScope{}, mapGlobalPreparationError(err)
	}
	if err := database.BindGlobalSearchPath(ctx); err != nil {
		return GlobalScope{}, mapGlobalPreparationError(err)
	}
	if err := database.VerifyGlobalEnvironmentID(ctx, environmentID); err != nil {
		return GlobalScope{}, mapGlobalPreparationError(err)
	}
	groupGeneration, billingGeneration := request.generations()
	if err := verifyGlobalBillingAuthority(
		ctx, database, environmentID, request.mode(), groupGeneration, billingGeneration,
	); err != nil {
		return GlobalScope{}, mapGlobalPreparationError(err)
	}
	state := &globalScopeState{
		actorID: request.actorID(), environmentID: environmentID,
		operationID: request.operationID(), mode: request.mode(),
		groupGeneration: groupGeneration, billingGeneration: billingGeneration,
		transaction: database,
	}
	state.active.Store(true)
	return GlobalScope{state: state}, nil
}

func verifyGlobalBillingAuthority(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	environmentID uuid.UUID,
	mode ScopeMode,
	groupGeneration int64,
	billingGeneration int64,
) error {
	var members []string
	var err error
	if mode == Command {
		if err = database.LockGlobalBillingGroup(
			ctx, environmentID, groupGeneration,
		); err == nil {
			members, err = database.LockGlobalBillingMembers(ctx, environmentID)
		}
		if err == nil {
			err = database.LockGlobalBillingDomain(ctx, environmentID, billingGeneration)
		}
	} else if mode == SnapshotQuery {
		if err = database.VerifyGlobalBillingGroup(
			ctx, environmentID, groupGeneration,
		); err == nil {
			members, err = database.VerifyGlobalBillingMembers(ctx, environmentID)
		}
		if err == nil {
			err = database.VerifyGlobalBillingDomain(ctx, environmentID, billingGeneration)
		}
	} else {
		return postgres.ErrAuthorityMismatch
	}
	if err != nil || !slices.Equal(members, []string{"billing"}) {
		return postgres.ErrAuthorityMismatch
	}
	return nil
}
