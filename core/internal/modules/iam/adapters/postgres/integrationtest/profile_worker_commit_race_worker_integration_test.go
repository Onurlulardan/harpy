//go:build integration

package integrationtest

import (
	"context"

	maintenancepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/maintenance"
	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type publicationPreparedWorker struct {
	transaction pgx.Tx
	outcome     maintenanceapp.Outcome
}

func (worker *publicationPreparedWorker) Commit(ctx context.Context) error {
	return worker.transaction.Commit(ctx)
}

func (worker *publicationPreparedWorker) Rollback(ctx context.Context) {
	_ = worker.transaction.Rollback(ctx)
}

func preparePublicationProfileWorker(
	ctx context.Context, db publicationDatabase, planID uuid.UUID,
) (*publicationPreparedWorker, error) {
	claim, err := publicationWorkerRaceClaim(ctx, db, planID)
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
	store := maintenancepostgres.NewStore(
		runtime, maintenancepostgres.DatabaseAccess{
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
		},
	)
	commands, err := maintenanceapp.NewCommands(
		store, func() (uuid.UUID, error) {
			return uuid.MustParse(
				"019c6a7e-7c00-7000-8000-000000000502",
			), nil
		},
	)
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

func publicationWorkerRaceClaim(
	ctx context.Context, db publicationDatabase, planID uuid.UUID,
) (maintenanceapp.Claim, error) {
	var jobID uuid.UUID
	err := db.admin.QueryRow(ctx, `SELECT job_id
FROM t_profile_publication.iam_profile_backfill_pages
WHERE plan_id=$1 AND page_number=1`, planID).Scan(&jobID)
	if err != nil {
		return maintenanceapp.Claim{}, err
	}
	reference, digest, err := maintenanceapp.ProfileReference(
		uuid.MustParse(db.tenant.id), planID, 1,
	)
	if err != nil {
		return maintenanceapp.Claim{}, err
	}
	return maintenanceapp.Claim{
		JobID: jobID,
		AttemptID: uuid.MustParse(
			"56000000-0000-4000-8000-000000000002",
		),
		TenantID:         uuid.MustParse(db.tenant.id),
		Kind:             maintenanceapp.KindProfileBackfill,
		PayloadRevision:  maintenanceapp.PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest,
	}, nil
}
