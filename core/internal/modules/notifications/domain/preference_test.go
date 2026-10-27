package domain

import (
	"testing"
	"time"
)

func TestMandatoryAndExplicitOptInPolicies(t *testing.T) {
	recipient, _ := ParseID("11111111-1111-4111-8111-111111111111")
	security, _ := DefaultPreference(recipient, CategorySecurity, ChannelEmail)
	security.Enabled = false
	decision, err := security.Decide(time.Now(), false)
	if err != nil || !decision.Enabled || !decision.Mandatory {
		t.Fatalf("security decision = %+v, %v", decision, err)
	}
	telegram, _ := DefaultPreference(recipient, CategoryWorkflowAction, ChannelTelegram)
	telegram.Enabled = true
	decision, _ = telegram.Decide(time.Now(), false)
	if decision.Enabled {
		t.Fatal("Telegram without explicit verified target was enabled")
	}
}

func TestQuietHoursMaterializesAcrossDST(t *testing.T) {
	quiet := QuietHours{TimeZone: "Europe/Istanbul", StartMinute: 22 * 60, EndMinute: 8 * 60}
	now := time.Date(2026, 7, 18, 21, 30, 0, 0, time.UTC)
	got, err := quiet.MaterializeNotBefore(now)
	want := time.Date(2026, 7, 19, 5, 0, 0, 0, time.UTC)
	if err != nil || !got.Equal(want) {
		t.Fatalf("MaterializeNotBefore() = %s, %v", got, err)
	}
}
