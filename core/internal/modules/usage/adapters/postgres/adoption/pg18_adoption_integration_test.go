//go:build integration

package adoption

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

func TestUsageAdoptionOnFreshAndAdoptedPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database := startUsageDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetAndApplyUsage(t, ctx, database, origin)
			exerciseLegacyCrashResume(t, ctx, database)
			exerciseManifestDriftFence(t, ctx, database)
			exerciseZeroSource(t, ctx, database)
		})
	}
}

func createLegacyTraceSource(t *testing.T, ctx context.Context, database *sql.DB) (
	*LegacySource, uuid.UUID,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `DROP SCHEMA IF EXISTS legacy_usage CASCADE;
CREATE SCHEMA legacy_usage;
CREATE TABLE legacy_usage.trace_logs(
id uuid PRIMARY KEY,tenant_id uuid,user_id uuid,method text,path text,status_code integer,
duration_ms bigint,start_time timestamptz,ip_address text,user_agent text,
request_body text,response_body text);
INSERT INTO legacy_usage.trace_logs VALUES(
'30000000-0000-4000-8000-000000000003',NULL,
'60000000-0000-4000-8000-000000000006','GET',
'/api/forms/550e8400-e29b-41d4-a716-446655440000?token=secret',200,3,
'2026-08-01 11:00:00+00','127.0.0.1','secret-agent','secret-body','secret-response')`)
	if err != nil {
		t.Fatal(err)
	}
	environment, defaultTenant := uuid.New(), uuid.New()
	source, err := NewLegacySourceForProfile(sourceRunner(database),
		"legacy_usage", "trace_logs", environment, defaultTenant)
	if err != nil {
		t.Fatal(err)
	}
	return source, defaultTenant
}

func exerciseLegacyCrashResume(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	source, _ := createLegacyTraceSource(t, ctx, database)
	store, err := NewStoreForAdopter(targetRunner(database))
	if err != nil {
		t.Fatal(err)
	}
	target := &crashTarget{Target: store, fail: true}
	clock := integrationClock{time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}
	coordinator, _ := usageapp.NewCoordinator(source, target, clock)
	request := usageapp.ExecuteRequest{OperationID: uuid.New(), RunID: uuid.Must(uuid.NewV7()),
		LeaseOwner: uuid.New(), PageLimit: 100, PageBudget: 10}
	_, err = coordinator.Execute(ctx, request)
	if !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("expected apply/advance crash, got %v", err)
	}
	run, err := coordinator.Execute(ctx, request)
	if err != nil || run.Status != usageapp.StatusCompleted || run.CopiedCount != 1 {
		t.Fatalf("adoption resume failed: run=%+v err=%v", run, err)
	}
	var route string
	var facts, mappings, receipts int
	err = database.QueryRowContext(ctx, `SELECT min(route_pattern),count(*) FROM platform.usage_request_facts`).
		Scan(&route, &facts)
	if err != nil || route != "/api/forms/{id}" || facts != 1 {
		t.Fatalf("adopted facts route=%q count=%d err=%v", route, facts, err)
	}
	_ = database.QueryRowContext(ctx, `SELECT count(*) FROM platform.usage_adoption_mappings`).Scan(&mappings)
	_ = database.QueryRowContext(ctx, `SELECT count(*) FROM platform.usage_adoption_page_receipts`).Scan(&receipts)
	if mappings != 1 || receipts != 1 {
		t.Fatalf("adoption evidence mappings=%d receipts=%d", mappings, receipts)
	}
	var rawLeak int
	err = database.QueryRowContext(ctx, `SELECT count(*) FROM platform.usage_request_facts
WHERE route_pattern LIKE '%?%' OR route_pattern LIKE '%550e8400%'`).Scan(&rawLeak)
	if err != nil || rawLeak != 0 {
		t.Fatalf("raw path leaked count=%d err=%v", rawLeak, err)
	}
	var tableText string
	_ = database.QueryRowContext(ctx, `SELECT string_agg(route_pattern,'') FROM platform.usage_request_facts`).Scan(&tableText)
	if strings.Contains(tableText, "secret") {
		t.Fatalf("sensitive legacy values leaked: %s", tableText)
	}
}

func exerciseManifestDriftFence(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	source, _ := createLegacyTraceSource(t, ctx, database)
	manifest, err := source.Inspect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	store, _ := NewStoreForAdopter(targetRunner(database))
	mutations := []func(*usageapp.Manifest){
		func(value *usageapp.Manifest) { value.RelationOID++ },
		func(value *usageapp.Manifest) { value.RowCount++ },
		func(value *usageapp.Manifest) { value.RootSHA256[0] ^= 0xff },
	}
	for index, mutate := range mutations {
		operation, runID := uuid.New(), uuid.Must(uuid.NewV7())
		run, _, startErr := store.Start(ctx, usageapp.StartCommand{OperationID: operation,
			RunID: runID, Manifest: manifest, TransformSHA256: usageapp.TransformSHA256,
			ObservedAt: time.Date(2026, 8, 1, 13, 0, index, 0, time.UTC)})
		if startErr != nil {
			t.Fatal(startErr)
		}
		drifted := manifest
		mutate(&drifted)
		fenced, claimErr := store.Claim(ctx, usageapp.ClaimCommand{OperationID: operation,
			RunID: runID, LeaseOwner: uuid.New(), ExpectedRevision: run.Revision,
			Manifest: drifted, ObservedAt: time.Date(2026, 8, 1, 13, 1, index, 0, time.UTC)})
		if claimErr != nil || fenced.Status != usageapp.StatusFenced {
			t.Fatalf("manifest drift %d not fenced: run=%+v err=%v", index, fenced, claimErr)
		}
	}
}

func exerciseZeroSource(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	source, _ := usageapp.NewZeroSource("zero-snapshot-0001")
	store, _ := NewStoreForAdopter(targetRunner(database))
	coordinator, _ := usageapp.NewCoordinator(source, store,
		integrationClock{time.Date(2026, 8, 1, 14, 0, 0, 0, time.UTC)})
	run, err := coordinator.Execute(ctx, usageapp.ExecuteRequest{OperationID: uuid.New(),
		RunID: uuid.Must(uuid.NewV7()), LeaseOwner: uuid.New(), PageLimit: 100, PageBudget: 1})
	if err != nil || run.Status != usageapp.StatusCompleted || run.CopiedCount != 0 {
		t.Fatalf("zero-source DB adoption failed: run=%+v err=%v", run, err)
	}
}
