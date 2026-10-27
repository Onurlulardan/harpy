//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"fmt"
	"runtime/debug"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func createRecoveryRootPlan(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime, plan profilebackfillapp.Plan, operationID uuid.UUID,
) profilebackfillapp.Result {
	t.Helper()
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	plans, err := iammodule.NewProfileBackfillPlanRuntime(catalogs)
	if err != nil {
		t.Fatal(err)
	}
	tenantID := uuid.MustParse(db.tenant.id)
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := platformtx.AdmitCommandScope(
		uuid.MustParse(publicationActorID), tenantID,
		"iam.profile.publish", "iam_system_resources",
		generations["iam"], generations,
	)
	if err != nil {
		t.Fatal(err)
	}
	evidence := profilebackfillapp.Evidence{
		OperationID:          operationID,
		RequestSHA256:        sha256.Sum256([]byte("recovery-root:" + plan.ID.String())),
		ActorID:              uuid.MustParse(publicationActorID),
		ManagePermissionID:   uuid.MustParse(publicationManagePerm),
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
	var result profilebackfillapp.Result
	err = runtime.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = plans.Create(scoped, scope, profilebackfillapp.CreateCommand{
				Plan: plan, Evidence: evidence,
			})
			return err
		},
	)
	if err != nil || result.PlanID != plan.ID || result.Replayed {
		t.Fatalf("create recovery root plan = %#v/%v\nCAS: %s\n%s",
			result, err, readRecoveryRootCASState(ctx, db, plan), debug.Stack())
	}
	return result
}

func readRecoveryRootCASState(
	ctx context.Context, db publicationDatabase, plan profilebackfillapp.Plan,
) string {
	schema := pgx.Identifier{db.tenant.schema}.Sanitize()
	statement := fmt.Sprintf(`SELECT
COALESCE((SELECT pg_catalog.string_agg(id::text||'/'||version::text||'/'||state,','
 ORDER BY version) FROM %s.iam_profile_schema_versions
 WHERE resource_kind=$1 AND state='active'),'none'),
COALESCE((SELECT pg_catalog.string_agg(id::text||'/'||version::text||'/'||state,','
 ORDER BY version) FROM %s.iam_profile_schema_versions
 WHERE resource_kind=$1 AND state='draft'),'none'),
COALESCE((SELECT pg_catalog.string_agg(id::text||'/'||source_schema_id::text||'/'||
 target_schema_id::text||'/'||state,',' ORDER BY created_at)
 FROM %s.iam_profile_backfill_plans
 WHERE resource_kind=$1 AND state IN ('planned','running','verified')),'none')`,
		schema, schema, schema)
	var active, drafts, plans string
	err := db.admin.QueryRow(ctx, statement, plan.Target.ResourceKind).Scan(
		&active, &drafts, &plans,
	)
	return fmt.Sprintf(
		"resource=%s source=%s target=%s version=%d active=[%s] drafts=[%s] plans=[%s] query=%v",
		plan.Target.ResourceKind, plan.SourceSchemaID, plan.Target.ID,
		plan.Target.Version, active, drafts, plans, err,
	)
}
