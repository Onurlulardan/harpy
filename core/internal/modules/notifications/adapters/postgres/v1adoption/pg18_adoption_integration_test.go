//go:build integration

package v1adoption

import (
	"context"
	"crypto/sha256"
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
	"github.com/google/uuid"
)

func TestNotificationsPG18BothOriginsCrashResumeIsolationAndImmutableAttempt(t *testing.T) {
	ctx := context.Background()
	connection := startNotificationPG18(t, ctx)
	prepareNotificationDatabase(t, ctx, connection)
	prepareNotificationTenant(t, ctx, connection, "fresh", "notification_fresh", false)
	prepareNotificationTenant(t, ctx, connection, "v1_adopted", "notification_a", true)
	prepareNotificationTenant(t, ctx, connection, "v1_adopted", "notification_b", true)
	var freshTemplates int
	if err := connection.QueryRow(ctx, `SELECT count(*) FROM notification_fresh.notifications_templates`).Scan(&freshTemplates); err != nil || freshTemplates != 1 {
		t.Fatalf("fresh template count=%d err=%v", freshTemplates, err)
	}
	tenantA := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	user := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	seedV1Notifications(t, ctx, connection, "notification_a", tenantA, user)
	_, err := connection.Exec(ctx, `UPDATE notification_a.notifications_writer_authority
SET writer='v1', lifecycle_state='writable', generation=1, row_version=2
WHERE singleton`)
	if err != nil {
		t.Fatal(err)
	}
	plan := app.Plan{RunID: uuid.MustParse("10000000-0000-4000-8000-000000000003"),
		TenantID: tenantA, SourceGeneration: "v1-main",
		SourceManifestDigest: sha256.Sum256([]byte("v1-notification-manifest")), PageSize: 1}
	first := executeNotificationStep(t, ctx, connection, "notification_a", plan, true)
	if first.Complete || first.AdoptedCount != 1 {
		t.Fatalf("first=%+v", first)
	}
	_ = executeNotificationStep(t, ctx, connection, "notification_a", plan, false)
	assertNotificationCounts(t, ctx, connection, 1, 1)
	second := executeNotificationStep(t, ctx, connection, "notification_a", plan, true)
	if second.Complete || second.AdoptedCount != 2 {
		t.Fatalf("second=%+v", second)
	}
	third := executeNotificationStep(t, ctx, connection, "notification_a", plan, true)
	if !third.Complete || third.AdoptedCount != 2 {
		t.Fatalf("third=%+v", third)
	}
	replay := executeNotificationStep(t, ctx, connection, "notification_a", plan, true)
	if !replay.Complete || replay.AdoptedCount != 2 {
		t.Fatalf("replay=%+v", replay)
	}
	assertNotificationCounts(t, ctx, connection, 2, 1)
	var tenantBMessages int
	if err = connection.QueryRow(ctx, `SELECT count(*) FROM notification_b.notifications_messages`).Scan(&tenantBMessages); err != nil || tenantBMessages != 0 {
		t.Fatalf("tenant B messages=%d err=%v", tenantBMessages, err)
	}
	assertAdoptedEvidence(t, ctx, connection)
	assertAttemptImmutableForAdopter(t, ctx, connection)
}
