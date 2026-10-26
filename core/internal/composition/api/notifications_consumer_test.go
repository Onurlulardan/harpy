package api

import (
	"testing"
	"time"
)

func TestNotificationsConsumerManifestRejectsExpiredOrNoncanonicalProof(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	expired := []byte(`{"version":1,"consumer":"frontend-v2","route":"/ws/notifications/{user_id}","replacement":"/ws/notifications","expires_at":"2026-07-19T11:59:59Z"}`)
	if _, err := parseNotificationConsumerManifest(expired, now); err == nil {
		t.Fatal("expired consumer manifest accepted")
	}
	noncanonical := append([]byte(" "), expired...)
	if _, err := parseNotificationConsumerManifest(noncanonical, now); err == nil {
		t.Fatal("noncanonical consumer manifest accepted")
	}
}
