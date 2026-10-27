package postgres

import (
	"context"
	"fmt"
	"reflect"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (repository *workerRepository) ProcessDatabaseJob(
	ctx context.Context,
	claim JobClaim,
	policy jobsv1.KindPolicy,
	binding platformasync.DatabaseJobBinding,
) error {
	tenantID, tenantScoped := claim.TenantID()
	if ctx == nil || !repository.valid() || repository.coordinator == nil ||
		!claim.valid(repository.identity) || !tenantScoped ||
		claim.EffectClass() != jobsv1.EffectDatabaseTransactional ||
		policy.Validate() != nil || policy.Kind != claim.Kind() ||
		policy.PayloadRevision != claim.PayloadRevision() ||
		policy.OwnerDomain != claim.authority.domain ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		policy.MaximumAttempts != claim.MaximumAttempts() ||
		binding.Owner() != claim.authority.domain {
		return ErrJobClaimInvalid
	}
	inTenant := repository.coordinator.InJobTenant
	if binding.RepeatableRead() {
		inTenant = repository.coordinator.InRepeatableReadJobTenant
	}
	return inTenant(
		ctx, tenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			if !platformtx.ContextOwnsJobScope(scoped, scope) ||
				platformtx.NetworkAllowed(scoped) {
				return ErrJobClaimInvalid
			}
			platformStore, err := platformtx.BindJob(
				scope, repository.databaseJobBinder,
			)
			if err != nil || platformStore == nil {
				return ErrWorkerStoreUnavailable
			}
			if err := platformStore.assertAuthority(scoped, claim); err != nil {
				return err
			}
			command, err := binding.Bind(scope)
			if err != nil || nilDatabaseJobCommand(command) {
				return ErrJobClaimInvalid
			}
			outcome, err := command.Execute(scoped, claim)
			if err != nil {
				return fmt.Errorf("execute owner database job command: %w", err)
			}
			if !outcome.Valid() {
				return platformasync.ErrDatabaseJobOutcomeInvalid
			}
			if successor, ok := outcome.Successor(); ok {
				if err := platformStore.enqueueSuccessor(
					scoped, claim, policy, successor,
				); err != nil {
					return err
				}
			}
			for _, child := range outcome.Children() {
				if err := platformStore.enqueueChild(scoped, claim, child); err != nil {
					return err
				}
			}
			for _, cancellation := range outcome.Cancellations() {
				if err := platformStore.cancelOwnedJob(
					scoped, claim, cancellation,
				); err != nil {
					return err
				}
			}
			for _, event := range outcome.Events() {
				if err := platformStore.enqueueEvent(scoped, claim, event); err != nil {
					return err
				}
			}
			return platformStore.complete(scoped, claim)
		},
	)
}

func nilDatabaseJobCommand(command platformasync.DatabaseJobCommand) bool {
	if command == nil {
		return true
	}
	value := reflect.ValueOf(command)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
