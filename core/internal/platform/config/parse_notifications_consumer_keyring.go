package config

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"unicode/utf8"
)

type notificationsConsumerKeyDocument struct {
	KeyID          string `json:"key_id"`
	Consumer       string `json:"consumer"`
	ManifestSHA256 string `json:"manifest_sha256"`
	PublicKey      string `json:"public_key"`
}

type notificationsConsumerKeyringDocument struct {
	Version int                                `json:"version"`
	Keys    []notificationsConsumerKeyDocument `json:"keys"`
}

func loadNotificationsConsumerKeyring(values environmentValues) (
	NotificationsConsumerKeyring, error,
) {
	path, found := values.lookup(notificationsConsumerKeyringFileKey)
	if !found {
		return NotificationsConsumerKeyring{}, configError(
			notificationsConsumerKeyringFileKey, "is required")
	}
	document, err := readSecretFile(notificationsConsumerKeyringFileKey, path, secretFileLimit)
	if err != nil {
		return NotificationsConsumerKeyring{}, err
	}
	return parseNotificationsConsumerKeyring([]byte(document))
}

func parseNotificationsConsumerKeyring(raw []byte) (NotificationsConsumerKeyring, error) {
	if !utf8.Valid(raw) || rejectDuplicateJSONMembers(raw) != nil {
		return NotificationsConsumerKeyring{}, invalidNotificationsConsumerKeyring()
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var document notificationsConsumerKeyringDocument
	if decoder.Decode(&document) != nil || document.Version != 1 ||
		len(document.Keys) < 1 || len(document.Keys) > 16 {
		return NotificationsConsumerKeyring{}, invalidNotificationsConsumerKeyring()
	}
	canonical, err := json.Marshal(document)
	if err != nil || !bytes.Equal(canonical, raw) {
		return NotificationsConsumerKeyring{}, invalidNotificationsConsumerKeyring()
	}
	keys := make(map[string]notificationsConsumerKey, len(document.Keys))
	previous := ""
	for _, item := range document.Keys {
		entry, valid := decodeNotificationsConsumerKey(item)
		if !valid || previous >= item.KeyID {
			return NotificationsConsumerKeyring{}, invalidNotificationsConsumerKeyring()
		}
		previous = item.KeyID
		keys[item.KeyID] = entry
	}
	keyring := NotificationsConsumerKeyring{keys: keys, valid: true}
	if keyring.Validate() != nil {
		return NotificationsConsumerKeyring{}, invalidNotificationsConsumerKeyring()
	}
	return keyring, nil
}

func decodeNotificationsConsumerKey(item notificationsConsumerKeyDocument) (
	notificationsConsumerKey, bool,
) {
	public, publicErr := base64.RawURLEncoding.Strict().DecodeString(item.PublicKey)
	digest, digestErr := base64.RawURLEncoding.Strict().DecodeString(item.ManifestSHA256)
	valid := catalogKeyIDPattern.MatchString(item.KeyID) &&
		notificationsConsumerPattern.MatchString(item.Consumer) && publicErr == nil &&
		digestErr == nil && len(public) == ed25519.PublicKeySize && len(digest) == 32 &&
		base64.RawURLEncoding.EncodeToString(public) == item.PublicKey &&
		base64.RawURLEncoding.EncodeToString(digest) == item.ManifestSHA256 && !allZero(digest)
	if !valid {
		clear(public)
		clear(digest)
		return notificationsConsumerKey{}, false
	}
	var fingerprint [32]byte
	copy(fingerprint[:], digest)
	clear(digest)
	return notificationsConsumerKey{consumer: item.Consumer, digest: fingerprint,
		key: append(ed25519.PublicKey(nil), public...)}, true
}
