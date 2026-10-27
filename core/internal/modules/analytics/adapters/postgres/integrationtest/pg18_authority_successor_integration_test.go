//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestAnalyticsCutoverSuccessorAuthorityPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "successor-"+origin)
			fixture := prepareAnalyticsSuccessorPG18(t, ctx, database, origin)
			seedAnalyticsWrongDomainSuccessor(t, ctx, database, fixture.digest)
			assertAnalyticsSuccessorPrivileges(t, ctx, database)
			assertAnalyticsSuccessorWrongPrincipal(t, ctx, database, fixture.digest)

			connection, err := database.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_, _ = connection.ExecContext(context.Background(),
					`RESET SESSION AUTHORIZATION`)
				_ = connection.Close()
			}()
			configureAnalyticsFinalizerConnection(t, ctx, connection)
			for _, test := range []struct {
				name, operationID, runID, digest string
			}{
				{"operation", analyticsSuccessorOperationB,
					analyticsSuccessorRunA, fixture.digest},
				{"run", analyticsSuccessorOperationA,
					analyticsSuccessorRunB, fixture.digest},
				{"digest", analyticsSuccessorOperationA,
					analyticsSuccessorRunA, fmt.Sprintf("%064x", 999)},
				{"domain", analyticsSuccessorWrongOp,
					analyticsSuccessorWrongRun, fixture.digest},
			} {
				_, err = callAnalyticsSuccessor(ctx, connection,
					test.operationID, test.runID, test.digest)
				requireAnalyticsPGCode(t, err, "P6101",
					"successor wrong "+test.name)
			}
			assertAnalyticsSuccessorGenerationMismatch(
				t, ctx, database, connection, fixture.digest,
			)
			assertAnalyticsSuccessorExtraMember(
				t, ctx, database, connection, fixture.digest,
			)
			first, err := callAnalyticsSuccessor(ctx, connection,
				analyticsSuccessorOperationA, analyticsSuccessorRunA,
				fixture.digest)
			if err != nil {
				t.Fatal(err)
			}
			second, err := callAnalyticsSuccessor(ctx, connection,
				analyticsSuccessorOperationA, analyticsSuccessorRunA,
				fixture.digest)
			if err != nil {
				t.Fatal(err)
			}
			if first != second {
				t.Fatalf("successor replay IDs differ: %s != %s", first, second)
			}
			var version, current int
			err = database.QueryRowContext(ctx, `
SELECT pg_catalog.uuid_extract_version($1::uuid),pg_catalog.count(*)
FROM platform.analytics_tenant_authorities
WHERE tenant_id=$2 AND authority_group_generation=2
GROUP BY 1`, first, analyticsTenantA).Scan(&version, &current)
			if err != nil {
				t.Fatal(err)
			}
			if version != 7 || current != 1 {
				t.Fatalf("successor version/current = %d/%d, want 7/1",
					version, current)
			}
		})
	}
}

func assertAnalyticsSuccessorPrivileges(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	var ownerExecute, finalizerExecute, finalizerInsert bool
	err := database.QueryRowContext(ctx, `
SELECT
pg_catalog.has_function_privilege('harpy_cutover_authority_owner',
'platform.analytics_finalize_authority_successor_v1(uuid,uuid,text)','EXECUTE'),
pg_catalog.has_function_privilege('harpy_cutover_finalizer',
'platform.analytics_finalize_authority_successor_v1(uuid,uuid,text)','EXECUTE'),
pg_catalog.has_table_privilege('harpy_cutover_finalizer',
'platform.analytics_tenant_authorities','INSERT')`,
	).Scan(&ownerExecute, &finalizerExecute, &finalizerInsert)
	if err != nil {
		t.Fatal(err)
	}
	if !ownerExecute || finalizerExecute || finalizerInsert {
		t.Fatalf("successor privileges owner/finalizer/table = %t/%t/%t",
			ownerExecute, finalizerExecute, finalizerInsert)
	}
}

func assertAnalyticsSuccessorWrongPrincipal(
	t *testing.T, ctx context.Context, database *sql.DB, digest string,
) {
	t.Helper()
	var authorityID string
	err := database.QueryRowContext(ctx, `
SELECT platform.analytics_finalize_authority_successor_v1($1,$2,$3)`,
		analyticsSuccessorOperationA, analyticsSuccessorRunA, digest,
	).Scan(&authorityID)
	requireAnalyticsPGCode(t, err, "42501", "successor wrong principal")
}

func assertAnalyticsSuccessorGenerationMismatch(
	t *testing.T, ctx context.Context, database *sql.DB,
	connection *sql.Conn, digest string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `UPDATE platform.cutover_groups
SET generation=3 WHERE scope_kind='tenant' AND scope_id=$1
AND group_id='analytics'`, analyticsTenantA); err != nil {
		t.Fatal(err)
	}
	_, err := callAnalyticsSuccessor(ctx, connection,
		analyticsSuccessorOperationA, analyticsSuccessorRunA, digest)
	requireAnalyticsPGCode(t, err, "P6101", "successor wrong generation")
	if _, err = database.ExecContext(ctx, `UPDATE platform.cutover_groups
SET generation=2 WHERE scope_kind='tenant' AND scope_id=$1
AND group_id='analytics'`, analyticsTenantA); err != nil {
		t.Fatal(err)
	}
}

func assertAnalyticsSuccessorExtraMember(
	t *testing.T, ctx context.Context, database *sql.DB,
	connection *sql.Conn, digest string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, `INSERT INTO
platform.cutover_group_members VALUES('tenant',$1,'analytics','forms',2)`,
		analyticsTenantA); err != nil {
		t.Fatal(err)
	}
	_, err := callAnalyticsSuccessor(ctx, connection,
		analyticsSuccessorOperationA, analyticsSuccessorRunA, digest)
	requireAnalyticsPGCode(t, err, "P6101", "successor extra member")
	if _, err = database.ExecContext(ctx, `DELETE FROM platform.cutover_group_members
WHERE scope_kind='tenant' AND scope_id=$1 AND group_id='analytics'
AND domain='forms'`, analyticsTenantA); err != nil {
		t.Fatal(err)
	}
}
