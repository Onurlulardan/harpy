package configfixture

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
	"path/filepath"
)

func setNotifications(fixture interface {
	EnvironmentSetter
	ArtifactFixture
}, worker bool) {
	fixture.Helper()
	configured := notificationsEntry(fixture, worker)
	fixture.Setenv(configured.name, configured.value)
}

func notificationsEnvironment(fixture ArtifactFixture, worker bool) string {
	fixture.Helper()
	configured := notificationsEntry(fixture, worker)
	return configured.name + "=" + configured.value
}

func notificationsEntry(fixture ArtifactFixture, worker bool) entry {
	directory := fixture.TempDir()
	if worker {
		path := filepath.Join(directory, "notifications-kek-keyring.json")
		document := `{"version":1,"current":{"revision":2,"key":"` +
			material(41) + `"},"retained":[]}`
		writeNotificationsFixture(fixture, path, document)
		return entry{"HARPY_NOTIFICATIONS_KEK_KEYRING_FILE", path}
	}
	path := filepath.Join(directory, "notifications-consumer-keyring.json")
	manifest := []byte(`{"version":1,"consumer":"frontend-v2","route":"/ws/notifications/{user_id}","replacement":"/ws/notifications","expires_at":"2030-01-01T00:00:00Z"}`)
	digest := sha256.Sum256(manifest)
	document := `{"version":1,"keys":[{"key_id":"frontend-v2-2026-07",` +
		`"consumer":"frontend-v2","manifest_sha256":"` +
		base64.RawURLEncoding.EncodeToString(digest[:]) +
		`","public_key":"` + material(31) + `"}]}`
	writeNotificationsFixture(fixture, path, document)
	return entry{"HARPY_NOTIFICATIONS_CONSUMER_KEYRING_FILE", path}
}

func writeNotificationsFixture(fixture ArtifactFixture, path, document string) {
	if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
		fixture.Fatalf("create Notifications keyring fixture: %v", err)
	}
}
