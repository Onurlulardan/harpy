package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"unicode/utf8"
)

type notificationsKEKDocument struct {
	Revision int64  `json:"revision"`
	Key      string `json:"key"`
}

type notificationsKEKKeyringDocument struct {
	Version  int                        `json:"version"`
	Current  notificationsKEKDocument   `json:"current"`
	Retained []notificationsKEKDocument `json:"retained"`
}

func loadNotificationsKEKKeyring(values environmentValues) (
	NotificationsKEKKeyring, error,
) {
	path, found := values.lookup(notificationsKEKKeyringFileKey)
	if !found {
		return NotificationsKEKKeyring{}, configError(
			notificationsKEKKeyringFileKey, "is required")
	}
	document, err := readSecretFile(notificationsKEKKeyringFileKey, path, secretFileLimit)
	if err != nil {
		return NotificationsKEKKeyring{}, err
	}
	return parseNotificationsKEKKeyring([]byte(document))
}

func parseNotificationsKEKKeyring(raw []byte) (NotificationsKEKKeyring, error) {
	if !utf8.Valid(raw) || rejectDuplicateJSONMembers(raw) != nil {
		return NotificationsKEKKeyring{}, invalidNotificationsKeyring()
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var document notificationsKEKKeyringDocument
	if decoder.Decode(&document) != nil || document.Version != 1 ||
		document.Retained == nil || len(document.Retained) > 7 {
		return NotificationsKEKKeyring{}, invalidNotificationsKeyring()
	}
	canonical, err := json.Marshal(document)
	if err != nil || !bytes.Equal(canonical, raw) {
		return NotificationsKEKKeyring{}, invalidNotificationsKeyring()
	}
	current, ok := decodeNotificationsKEK(document.Current, true)
	if !ok {
		return NotificationsKEKKeyring{}, invalidNotificationsKeyring()
	}
	keys := []NotificationsKEK{current}
	for _, item := range document.Retained {
		key, valid := decodeNotificationsKEK(item, false)
		if !valid {
			return NotificationsKEKKeyring{}, invalidNotificationsKeyring()
		}
		keys = append(keys, key)
	}
	keyring := NotificationsKEKKeyring{keys: keys, valid: true}
	if keyring.Validate() != nil {
		return NotificationsKEKKeyring{}, invalidNotificationsKeyring()
	}
	return keyring, nil
}

func decodeNotificationsKEK(document notificationsKEKDocument,
	current bool,
) (NotificationsKEK, bool) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(document.Key)
	if document.Revision < 1 || err != nil || len(decoded) != 32 ||
		base64.RawURLEncoding.EncodeToString(decoded) != document.Key || allZero(decoded) {
		clear(decoded)
		return NotificationsKEK{}, false
	}
	var material [32]byte
	copy(material[:], decoded)
	clear(decoded)
	return NotificationsKEK{revision: document.Revision,
		material: material, current: current}, true
}

func invalidNotificationsKeyring() error {
	return configError(notificationsKEKKeyringFileKey,
		"must satisfy notifications-kek-keyring.v1")
}
