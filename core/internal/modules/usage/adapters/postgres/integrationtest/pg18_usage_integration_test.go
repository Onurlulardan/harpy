//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUsageLifecycleOnFreshAndAdoptedPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database := startUsageDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetAndApplyUsage(t, ctx, database, origin)
			exerciseReplayAndConcurrency(t, ctx, database)
			exerciseTenantAndPlatformReads(t, ctx, database)
			exerciseAllReadShapes(t, ctx, database)
			exerciseRuntimeBoundary(t, ctx, database)
		})
	}
}

func exerciseReplayAndConcurrency(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	now := time.Date(2026, 8, 1, 12, 0, 0, 123456000, time.UTC)
	environment, tenant, actor := uuid.New(), uuid.New(), uuid.New()
	fact := newFact(environment, tenant, actor, "exact-replay", "/api/forms/{id}", now)
	id, replayed, err := recordFact(t, ctx, database, fact)
	if err != nil || replayed || id != fact.ID {
		t.Fatalf("first fact id=%s replay=%v err=%v", id, replayed, err)
	}
	retry := fact
	retry.ID = uuid.Must(uuid.NewV7())
	id, replayed, err = recordFact(t, ctx, database, retry)
	if err != nil || !replayed || id != fact.ID {
		t.Fatalf("fact replay id=%s replay=%v err=%v", id, replayed, err)
	}
	retry.Route = "/api/workflow/{id}"
	_, _, err = recordFact(t, ctx, database, retry)
	assertSQLState(t, err, "P6902")

	concurrent := newFact(environment, tenant, actor, "concurrent-replay",
		"/api/documents/{id}", now.Add(time.Microsecond))
	type result struct {
		id     uuid.UUID
		replay bool
		err    error
	}
	results := make(chan result, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			attempt := concurrent
			attempt.ID = uuid.Must(uuid.NewV7())
			id, replay, err := recordFact(t, ctx, database, attempt)
			results <- result{id, replay, err}
		}()
	}
	wait.Wait()
	close(results)
	seen := map[uuid.UUID]int{}
	replays := 0
	for result := range results {
		if result.err != nil {
			t.Fatal(result.err)
		}
		seen[result.id]++
		if result.replay {
			replays++
		}
	}
	if len(seen) != 1 || replays != 1 {
		t.Fatalf("concurrent replay drift: ids=%v replays=%d", seen, replays)
	}
}

func exerciseTenantAndPlatformReads(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	now := time.Date(2026, 8, 1, 13, 0, 0, 0, time.UTC)
	environment, tenantA, tenantB := uuid.New(), uuid.New(), uuid.New()
	actorA, actorB, principal := uuid.New(), uuid.New(), uuid.New()
	_, _, _ = recordFact(t, ctx, database,
		newFact(environment, tenantA, actorA, "tenant-a", "/api/forms/{id}", now))
	_, _, _ = recordFact(t, ctx, database,
		newFact(environment, tenantB, actorB, "tenant-b", "/api/workflow/{id}", now))
	var tenantJSON []byte
	err := database.QueryRowContext(ctx, `SELECT platform.usage_top_users_v1(
$1,$2,$2,$3,1,1,1,'tenant','saas','saas_customer',false,
'analytics.usage.read','usage.tenant.top_users',$4,$5,50)`, principal, tenantA,
		environment, now.Add(-time.Hour), now.Add(time.Hour)).Scan(&tenantJSON)
	if err != nil || !strings.Contains(string(tenantJSON), actorA.String()) ||
		strings.Contains(string(tenantJSON), actorB.String()) {
		t.Fatalf("tenant isolation failed: %s err=%v", tenantJSON, err)
	}
	var platformJSON []byte
	err = database.QueryRowContext(ctx, `SELECT platform.usage_top_users_v1(
$1,$2,NULL,$3,1,1,1,'platform','saas','saas_control',true,
'platform.analytics.usage.read','usage.platform.top_users',$4,$5,50)`, principal,
		tenantA, environment, now.Add(-time.Hour), now.Add(time.Hour)).Scan(&platformJSON)
	if err != nil || strings.Contains(string(platformJSON), actorA.String()) ||
		strings.Contains(string(platformJSON), actorB.String()) ||
		strings.Contains(string(platformJSON), tenantA.String()) {
		t.Fatalf("platform identity redaction failed: %s err=%v", platformJSON, err)
	}
	var subjects []struct {
		SubjectID string `json:"subject_id"`
	}
	if json.Unmarshal(platformJSON, &subjects) != nil || len(subjects) != 2 ||
		len(subjects[0].SubjectID) != 64 {
		t.Fatalf("platform pseudonym shape failed: %s", platformJSON)
	}
	err = database.QueryRowContext(ctx, `SELECT platform.usage_top_users_v1(
$1,$2,NULL,$3,1,1,1,'platform','saas','saas_customer',true,
'platform.analytics.usage.read','usage.platform.top_users',$4,$5,50)`, principal,
		tenantA, environment, now.Add(-time.Hour), now.Add(time.Hour)).Scan(&platformJSON)
	assertSQLState(t, err, "P6903")
}

func exerciseRuntimeBoundary(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx, `SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	_, err = connection.ExecContext(ctx, `SELECT fact_id FROM platform.usage_request_facts LIMIT 1`)
	assertSQLState(t, err, "42501")
	_, err = connection.ExecContext(ctx, `DELETE FROM platform.usage_request_facts`)
	assertSQLState(t, err, "42501")
	if _, err = connection.ExecContext(ctx, `RESET ROLE`); err != nil {
		t.Fatal(err)
	}
	var prohibited int
	err = database.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.columns
WHERE table_schema='platform' AND table_name='usage_request_facts'
AND column_name IN ('path','query','ip_address','user_agent','request_body','response_body','token')`).Scan(&prohibited)
	if err != nil || prohibited != 0 {
		t.Fatalf("prohibited raw storage columns=%d err=%v", prohibited, err)
	}
}
