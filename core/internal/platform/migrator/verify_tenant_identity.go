package migrator

import (
	"context"
	"database/sql"
	"fmt"

	postgresstore "github.com/Onurlulardan/harpy/core/internal/platform/migrator/adapters/postgres"
)

func (runner *Runner) verifyTenantIdentity(
	ctx context.Context,
	database *sql.DB,
	target postgresstore.TenantMigrationTarget,
	session postgresstore.TenantSessionProof,
) error {
	transaction, err := database.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("begin tenant identity verification: %w", err)
	}
	defer func() { _ = transaction.Rollback() }()
	if err := postgresstore.VerifyTenantMigrationSession(
		ctx, transaction, session, target,
	); err != nil {
		return err
	}
	origin := tenantSentinelOrigin(runner.tenantSelected[0].Origin)
	if err := postgresstore.VerifyTenantSchemaIdentity(
		ctx, transaction, target, origin,
	); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit tenant identity verification: %w", err)
	}
	return nil
}
