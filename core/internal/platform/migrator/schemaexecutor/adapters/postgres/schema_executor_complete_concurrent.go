package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func (store *schemaExecutorStore) completeConcurrentStep(
	ctx context.Context,
	session *schemaExecutorTenantSession,
	request schemaexecutor.ApplyRequest,
) error {
	operationCtx, cancel := schemaExecutorOperationContext(ctx)
	defer cancel()
	ctx = operationCtx
	transaction, err := session.connection.Begin(ctx)
	if err != nil {
		return schemaExecutorClassifyError(err)
	}
	defer store.rollbackTransaction(transaction)
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
