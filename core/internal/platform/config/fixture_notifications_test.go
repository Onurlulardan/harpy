package config

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"path/filepath"
)

func validNotificationsKeyringEnvironmentEntry() string {
	directory, err := os.MkdirTemp("", "harpy-notifications-keyring-")
	if err != nil {
		panic(err)
	}
	path := filepath.Join(directory, "keyring.json")
	key := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{41}, 32))
	document := `{"version":1,"current":{"revision":2,"key":"` + key +
		`"},"retained":[]}`
	if err = os.WriteFile(path, []byte(document), 0o600); err != nil {
		panic(err)
	}
	return notificationsKEKKeyringFileKey + "=" + path
}

func validNotificationsConsumerKeyringEnvironmentEntry() string {
	directory, err := os.MkdirTemp("", "harpy-notification-consumers-")
	if err != nil {
		panic(err)
	}
	path := filepath.Join(directory, "keyring.json")
	public := base64.RawURLEncoding.EncodeToString(
		bytes.Repeat([]byte{31}, ed25519.PublicKeySize),
	)
	manifest := []byte(`{"version":1,"consumer":"frontend-v2","route":"/ws/notifications/{user_id}","replacement":"/ws/notifications","expires_at":"2030-01-01T00:00:00Z"}`)
	digest := sha256.Sum256(manifest)
	document := `{"version":1,"keys":[{"key_id":"frontend-v2-2026-07",` +
		`"consumer":"frontend-v2","manifest_sha256":"` +
		base64.RawURLEncoding.EncodeToString(digest[:]) +
		`","public_key":"` + public + `"}]}`
	if err = os.WriteFile(path, []byte(document), 0o600); err != nil {
		panic(err)
	}
	return notificationsConsumerKeyringFileKey + "=" + path
}
