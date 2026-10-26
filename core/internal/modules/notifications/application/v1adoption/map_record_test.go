package v1adoption

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLegacyChannelAndCategoryMappingIsClosed(t *testing.T) {
	tenant := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	user := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	base := SourceRecord{ID: uuid.MustParse("10000000-0000-4000-8000-000000000003"),
		TenantID: tenant, UserID: &user, NotificationType: "password_reset",
		Title: "Reset", Body: "Body", Channel: "mail", CreatedAt: &now, InboxRevision: 1}
	mapped, issue, ignored := mapRecord(Plan{TenantID: tenant}, base)
	if issue != "" || ignored || mapped.Channel != "email" || mapped.Category != "security" {
		t.Fatalf("mail mapping=%+v issue=%s ignored=%t", mapped, issue, ignored)
	}
	base.Channel = "sms"
	if _, issue, _ = mapRecord(Plan{TenantID: tenant}, base); issue != "V1_CHANNEL_UNKNOWN" {
		t.Fatalf("unknown channel issue=%s", issue)
	}
}
