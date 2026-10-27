package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const notificationsKEKKeyringFileKey = "HARPY_NOTIFICATIONS_KEK_KEYRING_FILE"

var errInvalidNotificationsKEKKeyring = errors.New("Notifications KEK keyring is invalid")

type NotificationsKEK struct {
	revision int64
	material [32]byte
	current  bool
}

type NotificationsKEKKeyring struct {
	keys  []NotificationsKEK
	valid bool
}

func (keyring NotificationsKEKKeyring) Validate() error {
	if !keyring.valid || len(keyring.keys) < 1 || len(keyring.keys) > 8 {
		return errInvalidNotificationsKEKKeyring
	}
	current := 0
	revisions := make(map[int64]struct{}, len(keyring.keys))
	materials := make(map[[32]byte]struct{}, len(keyring.keys))
	for _, key := range keyring.keys {
		if key.revision < 1 || key.material == [32]byte{} {
			return errInvalidNotificationsKEKKeyring
		}
		if _, exists := revisions[key.revision]; exists {
			return errInvalidNotificationsKEKKeyring
		}
		if _, exists := materials[key.material]; exists {
			return errInvalidNotificationsKEKKeyring
		}
		revisions[key.revision] = struct{}{}
		materials[key.material] = struct{}{}
		if key.current {
			current++
		}
	}
	if current != 1 {
		return errInvalidNotificationsKEKKeyring
	}
	return nil
}

func (keyring NotificationsKEKKeyring) Keys() (map[int64][32]byte, error) {
	if err := keyring.Validate(); err != nil {
		return nil, err
	}
	result := make(map[int64][32]byte, len(keyring.keys))
	for _, key := range keyring.keys {
		result[key.revision] = key.material
	}
	return result, nil
}

func (keyring NotificationsKEKKeyring) CurrentRevision() (int64, error) {
	if err := keyring.Validate(); err != nil {
		return 0, err
	}
	for _, key := range keyring.keys {
		if key.current {
			return key.revision, nil
		}
	}
	return 0, errInvalidNotificationsKEKKeyring
}

func (NotificationsKEKKeyring) String() string {
	return "[REDACTED NOTIFICATIONS KEK KEYRING]"
}
func (NotificationsKEKKeyring) GoString() string {
	return "[REDACTED NOTIFICATIONS KEK KEYRING]"
}
func (NotificationsKEKKeyring) LogValue() slog.Value {
	return slog.StringValue("[REDACTED NOTIFICATIONS KEK KEYRING]")
}
func (NotificationsKEKKeyring) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED NOTIFICATIONS KEK KEYRING]")
}
