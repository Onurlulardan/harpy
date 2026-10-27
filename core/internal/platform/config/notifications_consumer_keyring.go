package config

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
)

const notificationsConsumerKeyringFileKey = "HARPY_NOTIFICATIONS_CONSUMER_KEYRING_FILE"

var notificationsConsumerPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

type notificationsConsumerKey struct {
	consumer string
	digest   [32]byte
	key      ed25519.PublicKey
}

type NotificationsConsumerKeyring struct {
	keys  map[string]notificationsConsumerKey
	valid bool
}

func (keyring NotificationsConsumerKeyring) Validate() error {
	if !keyring.valid || len(keyring.keys) < 1 || len(keyring.keys) > 16 {
		return invalidNotificationsConsumerKeyring()
	}
	for keyID, entry := range keyring.keys {
		if !catalogKeyIDPattern.MatchString(keyID) ||
			!notificationsConsumerPattern.MatchString(entry.consumer) ||
			entry.digest == [32]byte{} || len(entry.key) != ed25519.PublicKeySize {
			return invalidNotificationsConsumerKeyring()
		}
	}
	return nil
}

func (keyring NotificationsConsumerKeyring) Resolve(keyID, consumer string,
	digest [32]byte,
) (ed25519.PublicKey, error) {
	if keyring.Validate() != nil || !catalogKeyIDPattern.MatchString(keyID) ||
		!notificationsConsumerPattern.MatchString(consumer) || digest == [32]byte{} {
		return nil, errors.New("notification consumer key is unavailable")
	}
	entry, found := keyring.keys[keyID]
	if !found || entry.consumer != consumer || entry.digest != digest {
		return nil, errors.New("notification consumer key is unavailable")
	}
	return append(ed25519.PublicKey(nil), entry.key...), nil
}

func invalidNotificationsConsumerKeyring() error {
	return configError(notificationsConsumerKeyringFileKey,
		"must satisfy notifications-consumer-keyring.v1")
}

func (NotificationsConsumerKeyring) String() string {
	return "[REDACTED NOTIFICATIONS CONSUMER KEYRING]"
}
func (NotificationsConsumerKeyring) GoString() string {
	return "[REDACTED NOTIFICATIONS CONSUMER KEYRING]"
}
func (NotificationsConsumerKeyring) LogValue() slog.Value {
	return slog.StringValue("[REDACTED NOTIFICATIONS CONSUMER KEYRING]")
}
func (NotificationsConsumerKeyring) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED NOTIFICATIONS CONSUMER KEYRING]")
}
