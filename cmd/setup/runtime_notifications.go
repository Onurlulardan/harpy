package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
)

type notificationsKeyDocument struct {
	Revision int64  `json:"revision"`
	Key      string `json:"key"`
}

type notificationsKeyringDocument struct {
	Version  int                        `json:"version"`
	Current  notificationsKeyDocument   `json:"current"`
	Retained []notificationsKeyDocument `json:"retained"`
}

type consumerKeyDocument struct {
	KeyID          string `json:"key_id"`
	Consumer       string `json:"consumer"`
	ManifestSHA256 string `json:"manifest_sha256"`
	PublicKey      string `json:"public_key"`
}

type consumerKeyringDocument struct {
	Version int                   `json:"version"`
	Keys    []consumerKeyDocument `json:"keys"`
}

func writeNotificationsMaterial(dir string) ([]string, error) {
	kekMaterial, err := markedMaterial(15)
	if err != nil {
		return nil, err
	}
	kek, err := json.Marshal(notificationsKeyringDocument{
		Version:  1,
		Current:  notificationsKeyDocument{Revision: 1, Key: rawKey(kekMaterial[:])},
		Retained: make([]notificationsKeyDocument, 0),
	})
	if err != nil {
		return nil, err
	}
	kekPath, err := writeRuntimeFile(dir, "notifications-kek-keyring.json", kek)
	if err != nil {
		return nil, err
	}
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	manifest := sha256.Sum256([]byte("local-notifications-consumer"))
	consumer, err := json.Marshal(consumerKeyringDocument{
		Version: 1,
		Keys: []consumerKeyDocument{{
			KeyID: "frontend-v2-local", Consumer: "frontend-v2",
			ManifestSHA256: rawKey(manifest[:]), PublicKey: rawKey(public),
		}},
	})
	if err != nil {
		return nil, err
	}
	consumerPath, err := writeRuntimeFile(dir, "notifications-consumer-keyring.json", consumer)
	if err != nil {
		return nil, err
	}
	return []string{
		runtimeLine("HARPY_NOTIFICATIONS_KEK_KEYRING_FILE", kekPath),
		runtimeLine("HARPY_NOTIFICATIONS_CONSUMER_KEYRING_FILE", consumerPath),
	}, nil
}
