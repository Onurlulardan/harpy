//go:build integration

package maintenance

import (
	"context"
	"database/sql"
	"testing"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type maintenancePG18Rows struct{ *sql.Rows }

func (rows maintenancePG18Rows) Close() { _ = rows.Rows.Close() }

func executeMaintenancePG18Command(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
	claim maintenanceapp.Claim, nextID uuid.UUID, commit bool,
) (maintenanceapp.Outcome, error) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	defer connection.Close()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			return maintenanceapp.Outcome{}, err
		}
	}
	if _, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		schema+", platform, pg_catalog"); err != nil {
		_ = transaction.Rollback()
		return maintenanceapp.Outcome{}, err
	}
	var material [32]byte
	for index := range material {
		material[index] = byte(index + 1)
	}
	runtime, err := NewRuntime([]DirectoryKey{{
		ID: iamMaintenanceKeyID, Material: material, Active: true,
	}})
	if err != nil {
		_ = transaction.Rollback()
		return maintenanceapp.Outcome{}, err
	}
	defer runtime.Close()
	store := NewStore(runtime, DatabaseAccess{
		Query: func(ctx context.Context, query string, arguments ...any) (
			platformpostgres.Rows, error,
		) {
			rows, queryErr := transaction.QueryContext(ctx, query, arguments...)
			if queryErr != nil {
				return nil, queryErr
			}
			return maintenancePG18Rows{rows}, nil
		},
		QueryRow: func(
			ctx context.Context, query string, arguments ...any,
		) platformpostgres.Row {
			return transaction.QueryRowContext(ctx, query, arguments...)
		},
	})
	commands, err := maintenanceapp.NewCommands(store, func() (uuid.UUID, error) {
		return nextID, nil
	})
	if err != nil {
		_ = transaction.Rollback()
		return maintenanceapp.Outcome{}, err
	}
	outcome, err := commands.Execute(ctx, claim)
	if err != nil {
		_ = transaction.Rollback()
		return maintenanceapp.Outcome{}, err
	}
	if !commit {
		return outcome, transaction.Rollback()
	}
	return outcome, transaction.Commit()
}

func recurringMaintenancePG18Claim(
	t *testing.T, tenantID, jobID, attemptID string, kind maintenanceapp.Kind,
) maintenanceapp.Claim {
	t.Helper()
	tenant := uuid.MustParse(tenantID)
	reference, digest, err := maintenanceapp.RecurringReference(tenant, kind)
	if err != nil {
		t.Fatal(err)
	}
	return maintenanceapp.Claim{
		TenantID: tenant, JobID: uuid.MustParse(jobID),
		AttemptID: uuid.MustParse(attemptID), Kind: kind,
		PayloadRevision:  maintenanceapp.PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest,
	}
}
