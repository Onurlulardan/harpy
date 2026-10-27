package postgres

import (
	"context"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
)

func (store *schemaExecutorStore) applyConcurrentIndex(
	ctx context.Context,
	session *schemaExecutorTenantSession,
	request schemaexecutor.ApplyRequest,
) error {
	if err := schemaExecutorCheckLease(
		ctx, session.connection, request.Lease,
	); err != nil {
		return err
	}
	step := request.Plan.CurrentStep()
	prepared, err := schemaExecutorPrepareIndex(
		ctx, session.connection, session.schemaName, session.trigramSchema, step,
	)
	if err != nil {
		return schemaExecutorClassifyError(err)
	}
	if step.Operation() == schemav1.CreateIndexConcurrent && prepared.exists &&
		(!prepared.contract.valid || !prepared.contract.ready) {
		if err := store.dropInvalidIndex(ctx, session, request, step); err != nil {
			return err
		}
		return schemaexecutor.NewFault(
			schemaexecutor.FaultTransient, "schema_invalid_index_reconciled", nil,
		)
	}
	if prepared.apply {
		statement, err := ddl.Render(ddl.RenderContext{
			TenantSchema: session.schemaName, TrigramSchema: session.trigramSchema,
		}, step)
		if err != nil {
			return schemaExecutorUnexpectedOperation(step.Operation())
		}
		if _, err := session.connection.Exec(
			ctx, statement.Text, statement.Arguments...,
		); err != nil {
			return schemaExecutorClassifyError(err)
		}
	}
	if step.Operation() == schemav1.CreateIndexConcurrent {
		reconciled, err := schemaExecutorPrepareIndex(
			ctx, session.connection, session.schemaName, session.trigramSchema, step,
		)
		if err != nil {
			return schemaExecutorClassifyError(err)
		}
		if reconciled.exists &&
			(!reconciled.contract.valid || !reconciled.contract.ready) {
			if err := store.dropInvalidIndex(ctx, session, request, step); err != nil {
				return err
			}
			return schemaexecutor.NewFault(
				schemaexecutor.FaultTransient, "schema_invalid_index_reconciled", nil,
			)
		}
	}
	if err := schemaExecutorReconcileIndex(
		ctx, session.connection, session.schemaName, session.trigramSchema, step,
	); err != nil {
		return schemaExecutorClassifyError(err)
	}
	return store.completeConcurrentStep(ctx, session, request)
}

func (store *schemaExecutorStore) dropInvalidIndex(
	ctx context.Context,
	session *schemaExecutorTenantSession,
	request schemaexecutor.ApplyRequest,
	step schemaexecutor.ValidatedStep,
) error {
	cleanupStep, err := schemaExecutorCleanupStatement(session.schemaName, step)
	if err != nil {
		return err
	}
	if err := schemaExecutorCheckLease(
		ctx, session.connection, request.Lease,
	); err != nil {
		return err
	}
	_, err = session.connection.Exec(ctx, cleanupStep.Text)
	return schemaExecutorClassifyError(err)
}
