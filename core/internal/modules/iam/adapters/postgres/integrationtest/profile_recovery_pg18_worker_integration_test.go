//go:build integration

package integrationtest

import (
	"context"
	"fmt"
	"testing"

	maintenancepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/maintenance"
	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func executeRecoveryWorkerPage(
	t *testing.T, ctx context.Context, db publicationDatabase,
	planID uuid.UUID, page int,
) maintenanceapp.Outcome {
	t.Helper()
	worker, err := prepareRecoveryWorkerPage(ctx, db, planID, page)
	if err != nil {
		t.Fatalf("prepare recovery worker page %d: %v", page, err)
	}
	defer worker.Rollback(context.Background())
	if err = worker.Commit(ctx); err != nil {
		t.Fatalf("commit recovery worker page %d: %v", page, err)
	}
	return worker.outcome
}

func prepareRecoveryWorkerPage(
	ctx context.Context, db publicationDatabase, planID uuid.UUID, page int,
) (*publicationPreparedWorker, error) {
	claim, err := recoveryWorkerClaim(ctx, db, planID, page)
	if err != nil {
		return nil, err
	}
	transaction, err := db.admin.Begin(ctx)
	if err != nil {
		return nil, err
	}
	fail := func(cause error) (*publicationPreparedWorker, error) {
		_ = transaction.Rollback(ctx)
		return nil, cause
	}
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err = transaction.Exec(ctx, statement); err != nil {
			return fail(err)
		}
	}
	if _, err = transaction.Exec(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		db.tenant.schema+",platform,pg_catalog"); err != nil {
		return fail(err)
	}
	var material [32]byte
	for index := range material {
		material[index] = byte(index + 1)
	}
	runtime, err := maintenancepostgres.NewRuntime(
		[]maintenancepostgres.DirectoryKey{{
			ID: "AAAAAAAAAAAAAAAA", Material: material, Active: true,
		}},
	)
	if err != nil {
		return fail(err)
	}
	defer runtime.Close()
	store := maintenancepostgres.NewStore(runtime, maintenancepostgres.DatabaseAccess{
		Query: func(
			ctx context.Context, query string, arguments ...any,
		) (platformpostgres.Rows, error) {
			return transaction.Query(ctx, query, arguments...)
		},
		QueryRow: func(
			ctx context.Context, query string, arguments ...any,
		) platformpostgres.Row {
			return transaction.QueryRow(ctx, query, arguments...)
		},
	})
	commands, err := maintenanceapp.NewCommands(store, func() (uuid.UUID, error) {
		return uuid.MustParse("019c6a7e-7c00-7000-8000-000000000701"), nil
	})
	if err != nil {
		return fail(err)
	}
	outcome, err := commands.Execute(ctx, claim)
	if err != nil {
		return fail(err)
	}
	return &publicationPreparedWorker{
		transaction: transaction, outcome: outcome,
	}, nil
}

func recoveryWorkerClaim(
	ctx context.Context, db publicationDatabase, planID uuid.UUID, page int,
) (maintenanceapp.Claim, error) {
	var jobID uuid.UUID
	err := db.admin.QueryRow(ctx, `SELECT job_id
FROM t_profile_publication.iam_profile_backfill_pages
WHERE plan_id=$1 AND page_number=$2`, planID, page).Scan(&jobID)
	if err != nil {
		return maintenanceapp.Claim{}, err
	}
	reference, digest, err := maintenanceapp.ProfileReference(
		uuid.MustParse(db.tenant.id), planID, int64(page),
	)
	if err != nil {
		return maintenanceapp.Claim{}, err
	}
	attempt := uuid.NewSHA1(planID, []byte(fmt.Sprintf("attempt:%d", page)))
	return maintenanceapp.Claim{
		JobID: jobID, AttemptID: attempt,
		TenantID:         uuid.MustParse(db.tenant.id),
		Kind:             maintenanceapp.KindProfileBackfill,
		PayloadRevision:  maintenanceapp.PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest,
	}, nil
}
