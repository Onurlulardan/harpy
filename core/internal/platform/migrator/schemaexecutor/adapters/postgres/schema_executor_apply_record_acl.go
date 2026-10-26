package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/migrator/schemaexecutor/adapters/postgres/ddl"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5"
)

func schemaExecutorApplyRecordACL(
	ctx context.Context,
	executor pgx.Tx,
	schemaName string,
	step schemaexecutor.ValidatedStep,
) error {
	statements, err := ddl.RenderRecordACL(
		ddl.RenderContext{TenantSchema: schemaName}, step,
	)
	if err != nil {
		return err
	}
	for _, statement := range statements {
		if _, err := executor.Exec(ctx, statement.Text, statement.Arguments...); err != nil {
			return err
		}
	}
	return nil
}
