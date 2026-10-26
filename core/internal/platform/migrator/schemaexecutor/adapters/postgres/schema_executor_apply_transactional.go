package postgres

import (
	"context"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func (store *schemaExecutorStore) applyTransactional(
	ctx context.Context,
	session *schemaExecutorTenantSession,
	request schemaexecutor.ApplyRequest,
) error {
	transaction, err := session.connection.Begin(ctx)
	if err != nil {
		return schemaExecutorClassifyError(err)
	}
	defer store.rollbackTransaction(transaction)
	if err := schemaExecutorCheckLease(ctx, transaction, request.Lease); err != nil {
		return err
	}
	step := request.Plan.CurrentStep()
	if step.Operation() == schemav1.VerifyStorage {
		if err := schemaExecutorVerifyStorage(
			ctx, transaction, session.schemaName, session.trigramSchema, step,
		); err != nil {
			return schemaExecutorClassifyError(err)
		}
	} else {
		apply, err := schemaExecutorPrepareTransactional(
			ctx, transaction, session.schemaName, session.trigramSchema, step,
		)
		if err != nil {
			return schemaExecutorClassifyError(err)
		}
		if apply {
			statement, err := ddl.Render(ddl.RenderContext{
				TenantSchema: session.schemaName, TrigramSchema: session.trigramSchema,
			}, step)
			if err != nil {
				return schemaExecutorUnexpectedOperation(step.Operation())
			}
			if _, err := transaction.Exec(
				ctx, statement.Text, statement.Arguments...,
			); err != nil {
				return schemaExecutorClassifyError(err)
			}
			if err := schemaExecutorApplyRecordACL(
				ctx, transaction, session.schemaName, step,
			); err != nil {
				return schemaExecutorClassifyError(err)
			}
		}
	}
	return request.Lease.Synchronize(func(snapshot schemaexecutor.LeaseSnapshot) (
		schemaexecutor.LeaseSnapshot, error,
	) {
		updated, err := schemaExecutorCompleteStep(
			ctx, transaction, request, snapshot, "succeeded",
		)
		if err != nil {
			return snapshot, schemaExecutorClassifyError(err)
		}
		if err := transaction.Commit(ctx); err != nil {
			return snapshot, schemaExecutorClassifyError(err)
		}
		return updated, nil
	})
}
