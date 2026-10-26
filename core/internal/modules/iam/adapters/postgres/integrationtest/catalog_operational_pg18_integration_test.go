//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/jackc/pgx/v5"
)

const iamCatalogComponentID = "018f0000-0004-7000-8000-000000000001"

func TestIAMCatalogMaterializesOperationalRBACAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			db := startPublicationPG18(t, ctx, origin)
			preparePublicationCatalog(t, ctx, db)
			grantPublicationRuntimeAccess(t, ctx, db)
			seedPublicationV1Snapshot(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "tenant", "foundation"), true)
			seedIAMCatalogRoleSchema(t, ctx, db)
			seedPublicationSource(t, ctx, db, publicationSourceSchema(t))
			seedPublicationCurrentSource(t, ctx, db)
			seedPublicationEmptyMappingSource(t, ctx, db)
			applyPublicationDocuments(t, ctx, db,
				publicationIAMDocuments(t, origin, "tenant", "runtime"), true)
			seedPublicationAuthority(t, ctx, db)
			applyIAMCatalogDocuments(t, ctx, db, iamCatalogDocuments(t, origin))
			projectIAMCatalog(t, ctx, db)
			assertIAMCatalogOperationalReadback(t, ctx, db)
			activateLegacyIAMCatalog(t, ctx, db)
			applyIAMCatalogDocuments(
				t, ctx, db, iamCatalogVersionUpgradeDocuments(t, origin),
			)
			assertIAMCatalogTwoVersionUpgrade(t, ctx, db)
			applyIAMCatalogDocuments(
				t, ctx, db, iamCatalogLifecycleV3Documents(t, origin),
			)
			assertIAMCatalogLifecycleV3(t, ctx, db)
		})
	}
}

func iamCatalogDocuments(t *testing.T, origin string) []migration.Document {
	t.Helper()
	documents, err := iammodule.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	return selectPublicationDocuments(
		t, documents, origin, "tenant",
		func(version int64) bool {
			return version == 20260716014004 ||
				version == 20260716015700 || version == 20260716015701 ||
				version >= 20260716013430 && version <= 20260716013432 ||
				version >= 20260718180530 && version <= 20260718180532 ||
				version >= 20260730120000 && version <= 20260730120002 ||
				version >= 20260730142000 && version <= 20260730142009
		},
	)
}

func applyIAMCatalogDocuments(
	t *testing.T, ctx context.Context, db publicationDatabase,
	documents []migration.Document,
) {
	t.Helper()
	for _, document := range documents {
		tx, err := db.admin.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_owner")
		if err == nil {
			_, err = tx.Exec(
				ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
				db.tenant.schema+",platform,pg_catalog",
			)
		}
		if err == nil {
			_, err = tx.Exec(ctx, string(document.Content()))
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatalf("commit %s: %v", document.Path(), err)
		}
	}
}

func projectIAMCatalog(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	document := iamCatalogDocument(t)
	contentSHA256 := sha256.Sum256(document)
	requestSHA256 := sha256.Sum256([]byte("iam-catalog-operational-request-v1"))
	occurredAt := time.Date(2026, 7, 30, 10, 1, 0, 0, time.UTC)
	const statement = `SELECT projection.ready_at,
 iam_catalog_operational_readback_v1(projection.component_id)
FROM iam_catalog_project_v1(
 $1::name,$2::uuid,$3::uuid,$4,$5,$6::uuid,$7::uuid,$8,$9,$10,
 $11,$12,$13,$14,$15,$16,$17,$18,$19,$20
) AS projection`
	var readyAt time.Time
	var ready bool
	err := queryIAMCatalogTenant(t, ctx, db, statement,
		db.tenant.schema,
		"018f0000-0004-7000-8000-000000000003",
		"e3000000-0000-4000-8000-000000000001",
		"project", "1.0.0",
		"018f0000-0004-7000-8000-000000000004",
		iamCatalogComponentID, int64(1),
		catalogpackagev1.ContractIAM, hex.EncodeToString(contentSHA256[:]),
		"IAM_TEMPLATE", "project", "project", contentSHA256[:],
		2, 2, document, "project.iam.catalog.v1", requestSHA256[:], occurredAt,
	).Scan(&readyAt, &ready)
	if err != nil || !ready || !readyAt.Equal(occurredAt) {
		t.Fatalf("project IAM catalog = %s/%t/%v", readyAt, ready, err)
	}
}

func assertIAMCatalogOperationalReadback(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	const statement = `SELECT
 (SELECT count(*) FROM iam_catalog_operational_resources WHERE component_id=$1),
 (SELECT count(*) FROM iam_catalog_operational_grants WHERE component_id=$1),
 (SELECT count(DISTINCT role_id) FROM iam_catalog_operational_grants WHERE component_id=$1),
 (SELECT count(DISTINCT permission_id) FROM iam_catalog_operational_grants WHERE component_id=$1),
 (SELECT count(DISTINCT binding_id) FROM iam_catalog_operational_grants WHERE component_id=$1),
 (SELECT clock_value FROM iam_authorization_clock WHERE singleton),
 (SELECT authorization_version FROM iam_users WHERE id=$2)`
	var resources, grants, roles, permissions, bindings int
	var clock, authorization int64
	err := queryIAMCatalogTenant(t, ctx, db, statement,
		iamCatalogComponentID, publicationActorID,
	).Scan(
		&resources, &grants, &roles, &permissions, &bindings,
		&clock, &authorization,
	)
	if err != nil || resources != 2 || grants != 3 || roles != 1 ||
		permissions != 3 || bindings != 3 || clock != 1 ||
		authorization != 2 {
		t.Fatalf("IAM operational readback = %d/%d/%d/%d/%d/%d/%d/%v",
			resources, grants, roles, permissions, bindings,
			clock, authorization, err)
	}
}

func queryIAMCatalogTenant(
	t *testing.T, ctx context.Context, db publicationDatabase,
	statement string, arguments ...any,
) pgx.Row {
	t.Helper()
	tx, err := db.admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	if _, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
		"pg_catalog,"+db.tenant.schema+",platform"); err != nil {
		t.Fatal(err)
	}
	return &commitRow{Row: tx.QueryRow(ctx, statement, arguments...), tx: tx, t: t}
}

type commitRow struct {
	pgx.Row
	tx pgx.Tx
	t  *testing.T
}

func (row *commitRow) Scan(destinations ...any) error {
	err := row.Row.Scan(destinations...)
	if err == nil {
		err = row.tx.Commit(context.Background())
	}
	return err
}
