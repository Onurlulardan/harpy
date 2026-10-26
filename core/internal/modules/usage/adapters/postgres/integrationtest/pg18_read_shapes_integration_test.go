//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func exerciseAllReadShapes(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	now := time.Date(2026, 8, 1, 14, 0, 0, 0, time.UTC)
	environment, tenant, actor, principal := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, _, err := recordFact(t, ctx, database,
		newFact(environment, tenant, actor, "all-read-shapes", "/api/forms/{id}", now)); err != nil {
		t.Fatal(err)
	}
	start, end := now.Add(-time.Hour), now.Add(time.Hour)
	tenantPrefix := []any{principal, tenant, environment}
	queries := []struct {
		name, statement string
		arguments       []any
		contains        []string
	}{
		{"active", `SELECT platform.usage_active_users_v1($1,$2,$2,$3,1,1,1,
'tenant','saas','saas_customer',false,'analytics.usage.read','usage.tenant.active_users',$4)`,
			append(tenantPrefix, now), []string{`"dau"`, `"total_requests_today"`}},
		{"summary", `SELECT platform.usage_user_summary_v1($1,$2,$2,$3,1,1,1,
'tenant','saas','saas_customer',false,'analytics.usage.read','usage.tenant.user_summary',$4,$5,$6)`,
			append(tenantPrefix, actor.String(), start, end), []string{actor.String(), `"top_routes"`}},
		{"timeline", `SELECT platform.usage_user_timeline_v1($1,$2,$2,$3,1,1,1,
'tenant','saas','saas_customer',false,'analytics.usage.read','usage.tenant.user_timeline',$4,1,20)`,
			append(tenantPrefix, actor.String()), []string{`"items"`, `/api/forms/{id}`}},
		{"features", `SELECT platform.usage_features_v1($1,$2,$2,$3,1,1,1,
'tenant','saas','saas_customer',false,'analytics.usage.read','usage.tenant.features',$4,$5,20)`,
			append(tenantPrefix, start, end), []string{`"route_pattern"`, `"error_rate"`}},
		{"heatmap", `SELECT platform.usage_heatmap_v1($1,$2,$2,$3,1,1,1,
'tenant','saas','saas_customer',false,'analytics.usage.read','usage.tenant.heatmap',$4,$5)`,
			append(tenantPrefix, start, end), []string{`"day"`, `"hour"`}},
		{"trend", `SELECT platform.usage_trend_v1($1,$2,$2,$3,1,1,1,
'tenant','saas','saas_customer',false,'analytics.usage.read','usage.tenant.trend',$4,$5,'hour')`,
			append(tenantPrefix, start, end), []string{`"bucket"`, `"unique_users"`}},
		{"overview", `SELECT platform.usage_overview_v1($1,$2,$2,$3,1,1,1,
'tenant','saas','saas_customer',false,'analytics.usage.read','usage.tenant.overview',$4,$5,$6)`,
			append(tenantPrefix, start, end, now), []string{`"active_users"`, `"recent_trend"`}},
	}
	for _, test := range queries {
		t.Run(test.name, func(t *testing.T) {
			var document []byte
			err := database.QueryRowContext(ctx, test.statement, test.arguments...).Scan(&document)
			if err != nil || !json.Valid(document) {
				t.Fatalf("%s document=%s err=%v", test.name, document, err)
			}
			for _, expected := range test.contains {
				if !strings.Contains(string(document), expected) {
					t.Fatalf("%s missing %q: %s", test.name, expected, document)
				}
			}
		})
	}
	var document []byte
	err := database.QueryRowContext(ctx, `SELECT platform.usage_heatmap_v1($1,$2,$2,$3,1,1,1,
'tenant','saas','saas_customer',false,'analytics.usage.read','usage.tenant.trend',$4,$5)`,
		principal, tenant, environment, start, end).Scan(&document)
	assertSQLState(t, err, "P6903")
}
