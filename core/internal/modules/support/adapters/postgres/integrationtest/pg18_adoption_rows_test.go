//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
	"github.com/google/uuid"
)

func contactRow(id, name string, createdAt time.Time) application.ContactSourceRow {
	return application.ContactSourceRow{ID: uuid.MustParse(id), Name: name,
		Email: name + "@example.com", Company: "Forge", Phone: "+905555555555",
		EmployeeCount: "11-50", Message: "Legacy support request", ContactType: "general",
		Status: "new", Source: "legacy", IPAddress: "192.0.2.1",
		UserAgent: "legacy-browser", CreatedAt: createdAt}
}

func bugRow(id, tenantID uuid.UUID, createdAt time.Time) application.BugSourceRow {
	return application.BugSourceRow{ID: id, TenantID: tenantID,
		ReporterID: uuid.New(), ReporterName: "Ada", ReporterEmail: "ada@example.com",
		Title: "Legacy bug", Description: "Legacy report description",
		Category: "bug", Severity: "high", Priority: "normal",
		PageURL: "https://example.com/legacy", UserAgent: "legacy-browser",
		BrowserInfo: []byte(`{"name":"browser"}`), ConsoleErrors: []byte(`[]`),
		NetworkInfo: []byte(`[]`), Screenshots: []byte(`[{"legacy_id":"shot-1"}]`),
		StepsToReproduce: "Open the legacy page", Status: "open",
		CreatedAt: createdAt, UpdatedAt: createdAt}
}

func assertAdoptionPartitions(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var contacts, bugs, mappings, mismatches, screenshots int
	if err := database.QueryRowContext(ctx,
		`SELECT count(*) FROM platform.support_contact_messages`).Scan(&contacts); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx,
		`SELECT count(*) FROM platform.support_bug_reports`).Scan(&bugs); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, `SELECT count(*)
        FROM platform.support_adoption_mappings
        WHERE target_id<>source_id AND pg_catalog.uuid_extract_version(target_id)=7`).Scan(&mappings); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, `SELECT count(*)
        FROM platform.support_adoption_mappings mapping
        JOIN platform.support_bug_reports report ON report.id=mapping.target_id
        WHERE mapping.source_kind='bug_tenant'
          AND mapping.source_tenant_id IS DISTINCT FROM report.tenant_id`).Scan(&mismatches); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, `SELECT count(*)
        FROM platform.support_bug_reports
        WHERE pg_catalog.jsonb_array_length(legacy_screenshot_refs)=1`).Scan(&screenshots); err != nil {
		t.Fatal(err)
	}
	if contacts != 2 || bugs != 2 || mappings != 4 || mismatches != 0 || screenshots != 2 {
		t.Fatalf("adoption evidence contacts=%d bugs=%d mappings=%d mismatches=%d screenshots=%d",
			contacts, bugs, mappings, mismatches, screenshots)
	}
}

func assertAdopterPrivileges(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	_, err := connection.ExecContext(ctx, `SELECT run_id FROM platform.support_adoption_runs LIMIT 1`)
	assertSQLState(t, err, "42501")
	_, err = connection.ExecContext(ctx, `INSERT INTO platform.support_adoption_mappings(
        run_id,source_kind,source_id,target_id,source_sha256,target_sha256,page_number,applied_at)
        VALUES ('00000000-0000-0000-0000-000000000000','contact_global',
        '00000000-0000-0000-0000-000000000001',pg_catalog.uuidv7(),
        pg_catalog.sha256('x'::bytea),pg_catalog.sha256('y'::bytea),1,pg_catalog.now())`)
	assertSQLState(t, err, "42501")
}
