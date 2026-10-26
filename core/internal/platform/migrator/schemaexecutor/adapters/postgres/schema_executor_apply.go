package postgres

import (
	"context"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func (store *schemaExecutorStore) Apply(
	ctx context.Context,
	request schemaexecutor.ApplyRequest,
) error {
	session, err := store.openTenantSession(ctx, request.Lease)
	if err != nil {
		return schemaExecutorClassifyError(err)
	}
	defer session.close(store.cleanupTimeout)
	step := request.Plan.CurrentStep()
	switch step.TransactionClass() {
	case schemav1.Transactional:
		if step.Operation() == schemav1.BackfillFieldBatch {
			return store.applyBackfill(ctx, session, request)
		}
		return store.applyTransactional(ctx, session, request)
	case schemav1.ConcurrentIndex:
		return store.applyConcurrentIndex(ctx, session, request)
	default:
		return schemaexecutor.NewFault(
			schemaexecutor.FaultPermanent, "schema_transaction_class_invalid", nil,
		)
	}
}

func schemaExecutorCheckLease(
	ctx context.Context,
	querier schemaExecutorCatalogQuerier,
	lease schemaexecutor.Lease,
) error {
	var tenantID string
	operationCtx, cancel := schemaExecutorOperationContext(ctx)
	defer cancel()
	ctx = operationCtx
	err := lease.Synchronize(func(snapshot schemaexecutor.LeaseSnapshot) (
		schemaexecutor.LeaseSnapshot, error,
	) {
		err := querier.QueryRow(ctx, schemaExecutorCheckLeaseSQL,
			lease.JobID(), lease.LeaseID(), lease.ExecutorID(), lease.JobAttempt(),
			lease.PlanSHA256(), lease.StepOrdinal(), lease.StepSHA256(),
			snapshot.JobGeneration, snapshot.StepGeneration, lease.StepAttempt(),
		).Scan(&tenantID)
		return snapshot, err
	})
	if err != nil || tenantID != lease.TenantID() {
		return schemaExecutorLeaseLost()
	}
	return nil
}

func schemaExecutorUnexpectedOperation(operation schemav1.OperationID) error {
	return schemaexecutor.NewFault(
		schemaexecutor.FaultPermanent,
		"schema_operation_not_implemented",
		fmt.Errorf("operation %s has no executor", operation),
	)
}
