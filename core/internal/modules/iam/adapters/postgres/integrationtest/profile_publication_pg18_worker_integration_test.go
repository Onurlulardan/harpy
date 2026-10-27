//go:build integration

package integrationtest

import (
	"context"
	"testing"

	maintenancepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/maintenance"
	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func executePublicationProfileWorker(
	t *testing.T, ctx context.Context, db publicationDatabase, planID uuid.UUID,
) {
	t.Helper()
	var jobID uuid.UUID
	if err := db.admin.QueryRow(ctx, `SELECT job_id
FROM t_profile_publication.iam_profile_backfill_pages
WHERE plan_id=$1 AND page_number=1`, planID).Scan(&jobID); err != nil {
		t.Fatal(err)
	}
	reference, digest, err := maintenanceapp.ProfileReference(
		uuid.MustParse(db.tenant.id), planID, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	claim := maintenanceapp.Claim{
		JobID:     jobID,
		AttemptID: uuid.NewSHA1(planID, []byte("publication-attempt-1")),
		TenantID:  uuid.MustParse(db.tenant.id), Kind: maintenanceapp.KindProfileBackfill,
		PayloadRevision:  maintenanceapp.PayloadRevision,
		PayloadReference: reference, PayloadSHA256: digest,
	}
	tx, err := db.admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err = tx.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = tx.Exec(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		db.tenant.schema+",platform,pg_catalog"); err != nil {
		t.Fatal(err)
	}
	var material [32]byte
	for index := range material {
		material[index] = byte(index + 1)
	}
	runtime, err := maintenancepostgres.NewRuntime([]maintenancepostgres.DirectoryKey{{
		ID: "AAAAAAAAAAAAAAAA", Material: material, Active: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	store := maintenancepostgres.NewStore(runtime, maintenancepostgres.DatabaseAccess{
		Query: func(ctx context.Context, query string, arguments ...any) (
			platformpostgres.Rows, error,
		) {
			return tx.Query(ctx, query, arguments...)
		},
		QueryRow: func(
			ctx context.Context, query string, arguments ...any,
		) platformpostgres.Row {
			return tx.QueryRow(ctx, query, arguments...)
		},
	})
	commands, err := maintenanceapp.NewCommands(store, func() (uuid.UUID, error) {
		return uuid.MustParse("019c6a7e-7c00-7000-8000-000000000702"), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := commands.Execute(ctx, claim)
	if err != nil || outcome.Successor != nil {
		t.Fatalf("publication profile worker = %#v/%v", outcome, err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func assertPublicationWorkerActivation(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime, planID uuid.UUID,
) {
	t.Helper()
	var planState, sourceState, targetState string
	var completed, converted, failures int
	err := db.admin.QueryRow(ctx, `SELECT plan.state,source.state,target.state,
checkpoint.completed_pages,checkpoint.converted_count,checkpoint.error_count
FROM t_profile_publication.iam_profile_backfill_plans AS plan
JOIN t_profile_publication.iam_profile_schema_versions AS source
 ON source.id=plan.source_schema_id
JOIN t_profile_publication.iam_profile_schema_versions AS target
 ON target.id=plan.target_schema_id
JOIN t_profile_publication.iam_profile_backfill_checkpoints AS checkpoint
 ON checkpoint.plan_id=plan.id WHERE plan.id=$1`, planID).Scan(
		&planState, &sourceState, &targetState, &completed, &converted, &failures,
	)
	if err != nil || planState != "activated" || sourceState != "retired" ||
		targetState != "active" || completed != 1 || converted != 1 || failures != 0 {
		t.Fatalf("publication activation = %s/%s/%s/%d/%d/%d/%v",
			planState, sourceState, targetState, completed, converted, failures, err)
	}
	view, err := runtime.application.GetSelf(ctx, publicationClaims(t, db.tenant))
	if err != nil || publicationViewString(t, view.Document, "nickname") != "nick" ||
		publicationViewString(t, view.Document, "contact_alias_v2") == "" ||
		publicationViewString(t, view.Document, "display_name") != "" {
		t.Fatalf("activated profile view = %s/%v", view.Document, err)
	}
}
