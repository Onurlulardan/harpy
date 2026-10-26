package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func (store *schemaExecutorStore) applyBackfill(
	ctx context.Context,
	session *schemaExecutorTenantSession,
	request schemaexecutor.ApplyRequest,
) error {
	for {
		complete, err := store.applyBackfillBatch(ctx, session, request)
		if err != nil {
			return schemaExecutorClassifyError(err)
		}
		if complete {
			return nil
		}
	}
}

func (store *schemaExecutorStore) applyBackfillBatch(
	ctx context.Context,
	session *schemaExecutorTenantSession,
	request schemaexecutor.ApplyRequest,
) (bool, error) {
	operationCtx, cancel := schemaExecutorOperationContext(ctx)
	defer cancel()
	ctx = operationCtx
	transaction, err := session.connection.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer store.rollbackTransaction(transaction)
	complete := false
	err = request.Lease.Synchronize(func(snapshot schemaexecutor.LeaseSnapshot) (
		schemaexecutor.LeaseSnapshot, error,
	) {
		updated, done, err := store.executeBackfillBatch(
			ctx, transaction, session, request, snapshot,
		)
		complete = done
		return updated, err
	})
	return complete, err
}

func (store *schemaExecutorStore) executeBackfillBatch(
	ctx context.Context,
	transaction pgx.Tx,
	session *schemaExecutorTenantSession,
	request schemaexecutor.ApplyRequest,
	snapshot schemaexecutor.LeaseSnapshot,
) (schemaexecutor.LeaseSnapshot, bool, error) {
	progress, err := schemaExecutorLockLease(
		ctx, transaction, request.Lease, snapshot,
	)
	if err != nil {
		return snapshot, false, err
	}
	step := request.Plan.CurrentStep()
	if step.Operation() != schemav1.BackfillFieldBatch {
		return snapshot, false, schemaExecutorUnexpectedOperation(step.Operation())
	}
	tableName := step.Mapping().Form.PhysicalName
	if mapping, ok := step.V2Mapping(); ok {
		tableName = mapping.Form.PhysicalName
	}
	if err := schemaExecutorRequireTable(
		ctx, transaction, session.schemaName, tableName,
	); err != nil {
		return snapshot, false, err
	}
	if err := schemaExecutorValidateBackfillColumns(
		ctx, transaction, session.schemaName, step,
	); err != nil {
		return snapshot, false, err
	}
	statement, err := ddl.Render(ddl.RenderContext{
		TenantSchema: session.schemaName, TrigramSchema: session.trigramSchema,
	}, step)
	if err != nil {
		return snapshot, false, err
	}
	page, err := schemaExecutorRunBackfill(
		ctx, transaction, statement, progress.checkpointRecordID,
	)
	if err != nil {
		return snapshot, false, err
	}
	if page.count == 0 {
		updated, err := schemaExecutorCompleteStep(
			ctx, transaction, request, snapshot, "succeeded",
		)
		if err == nil {
			err = transaction.Commit(ctx)
		}
		return updated, err == nil, err
	}
	var stepGeneration int64
	err = transaction.QueryRow(ctx, schemaExecutorUpdateProgressSQL,
		request.Lease.JobID(), request.Lease.StepOrdinal(), request.Lease.StepAttempt(),
		request.Lease.PlanSHA256(), request.Lease.StepSHA256(), snapshot.StepGeneration,
		page.checkpoint, page.count, page.canonicalBytes, page.pageSHA256,
	).Scan(&stepGeneration)
	if err != nil {
		return snapshot, false, schemaExecutorLeaseRow(err, "advance backfill checkpoint")
	}
	snapshot.StepGeneration = stepGeneration
	if err := transaction.Commit(ctx); err != nil {
		return snapshot, false, err
	}
	return snapshot, false, nil
}
