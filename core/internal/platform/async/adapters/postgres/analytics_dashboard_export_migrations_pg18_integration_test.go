//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestAnalyticsDashboardExportMigrationsApplyOnPG184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			database := startAdoptionPG18(t, ctx, "analytics_export_"+origin)
			applyAnalyticsDashboardExportPG18Migrations(t, ctx, database, origin)
			var current, legacy bool
			err := database.QueryRowContext(ctx, `SELECT
    to_regprocedure('platform.platform_async_analytics_dashboard_export_unknown_resolve_v2(uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,uuid,text,bytea,timestamptz)') IS NOT NULL,
    pg_get_functiondef(to_regprocedure('platform.platform_async_analytics_dashboard_export_unknown_resolve_v1(uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,text,bytea,timestamptz)')) LIKE '%cleanup lease is required%'`).Scan(
				&current, &legacy,
			)
			if err != nil || !current || !legacy {
				t.Fatalf("resolve ABI state current=%v legacy_fenced=%v err=%v", current, legacy, err)
			}
		})
	}
}

func TestAnalyticsDashboardExportMaintenancePrefixThrough42RunsOnPG184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			database := startAdoptionPG18(t, ctx, "analytics_export_prefix_"+origin)
			applyAnalyticsDashboardExportPG18MigrationsThrough(
				t, ctx, database, origin, 20260802020042,
			)
			fixture := analyticsDashboardExportPG18{
				ctx: ctx, database: database,
				runtime: prepareAnalyticsDashboardExportPG18(t, ctx, database),
			}
			prepareAnalyticsMaintenancePG18(t, fixture)
			if _, err := database.ExecContext(ctx,
				analyticsDashboardExportMaintenanceRecoveryWrappersPG18); err != nil {
				t.Fatalf("prepare prefix maintenance wrappers: %v", err)
			}
			_, candidate := createAnalyticsMaintenanceUnknownCandidate(t, fixture,
				platformasync.AnalyticsDashboardExportDeleteJobKind, false)
			claim := claimAnalyticsMaintenanceUnknown(t, fixture, candidate)
			resolution := newAnalyticsMaintenanceResolution(t, candidate, "deleted")
			replayed, receiptID, state, err := callAnalyticsMaintenanceResolve(
				fixture, candidate, claim, resolution,
			)
			if err != nil || replayed || receiptID != resolution.effectReceiptID ||
				state != "succeeded" {
				t.Fatalf("prefix maintenance resolve replayed=%v receipt=%s state=%s err=%v",
					replayed, receiptID, state, err)
			}
			var identityExists bool
			if err = database.QueryRowContext(ctx, `SELECT to_regprocedure(
'platform.platform_async_analytics_export_maint_receipt_id_v1(uuid,text)'
) IS NOT NULL`).Scan(&identityExists); err != nil || identityExists {
				t.Fatalf("prefix unexpectedly depends on 020043: exists=%v err=%v",
					identityExists, err)
			}
		})
	}
}

func applyAnalyticsDashboardExportPG18Migrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	applyAnalyticsDashboardExportPG18MigrationsThrough(
		t, ctx, database, origin, 20260802020049,
	)
}

func applyAnalyticsDashboardExportPG18MigrationsThrough(
	t *testing.T, ctx context.Context, database *sql.DB, origin string, end int64,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, asyncACLRoleFixture); err != nil {
		t.Fatalf("create Analytics export roles: %v", err)
	}
	if _, err := database.ExecContext(ctx, adoptionAuthorityFixture); err != nil {
		t.Fatalf("create Analytics export authority fixture: %v", err)
	}
	if _, err := database.ExecContext(ctx, `ALTER TABLE platform.tenants
    ADD COLUMN schema_name name,
    ADD COLUMN status text,
    ADD COLUMN deleted_at timestamptz;`); err != nil {
		t.Fatalf("extend Analytics export tenant catalog: %v", err)
	}
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	selected := analyticsDashboardExportPG18Documents(t, documents, origin, end)
	for _, document := range selected {
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) {
				t.Fatalf("apply %s: %s position=%d detail=%s where=%s",
					document.Path(), postgresError.Message, postgresError.Position,
					postgresError.Detail, postgresError.Where)
			}
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
	}
}

func analyticsDashboardExportPG18Documents(
	t *testing.T, documents []migration.Document, origin string, end int64,
) []migration.Document {
	t.Helper()
	marker := "/migrations/" + origin + "/"
	selected := make([]migration.Document, 0, 100)
	for _, document := range documents {
		if !strings.Contains(document.Path(), marker) {
			continue
		}
		version := adoptionDocumentVersion(t, document.Path())
		if version <= 20260715022010 || version == 20260718020025 ||
			version >= 20260802020000 && version <= end {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return adoptionDocumentVersion(t, selected[left].Path()) <
			adoptionDocumentVersion(t, selected[right].Path())
	})
	rangeCount := 0
	for _, document := range selected {
		version := adoptionDocumentVersion(t, document.Path())
		if version >= 20260802020000 && version <= end {
			rangeCount++
		}
	}
	want := int(end-20260802020000) + 1
	if rangeCount != want {
		t.Fatalf("%s Analytics export migration range=%d, want %d through %d",
			origin, rangeCount, want, end)
	}
	return selected
}
