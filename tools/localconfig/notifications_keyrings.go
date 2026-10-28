package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"
)

type notificationsConsumerManifestDocument struct {
	Version     int    `json:"version"`
	Consumer    string `json:"consumer"`
	Route       string `json:"route"`
	Replacement string `json:"replacement"`
	ExpiresAt   string `json:"expires_at"`
}

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

type notificationsKEKDocument struct {
	Revision int    `json:"revision"`
	Key      string `json:"key"`
}

type notificationsKEKKeyringDocument struct {
	Version  int                        `json:"version"`
	Current  notificationsKEKDocument   `json:"current"`
	Retained []notificationsKEKDocument `json:"retained"`
}

func addNotificationsKeyrings(set *artifactSet, now time.Time) error {
	expiresAt := now.Add(90 * 24 * time.Hour)
	manifest, err := marshalJSON(notificationsConsumerManifestDocument{
		Version: 1, Consumer: "frontend-v2", Route: "/ws/notifications/{user_id}",
		Replacement: "/ws/notifications", ExpiresAt: canonicalTimestamp(expiresAt),
	})
	if err != nil {
		return err
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	seed := privateKey.Seed()
	defer clear(seed)
	defer clear(privateKey)
	keyID, err := randomLowerID("local-")
	if err != nil {
		return err
	}
	digest := sha256.Sum256(manifest)
	consumerRing := notificationsConsumerKeyringDocument{
		Version: 1,
		Keys: []notificationsConsumerKeyDocument{{
			KeyID: keyID, Consumer: "frontend-v2",
			ManifestSHA256: base64.RawURLEncoding.EncodeToString(digest[:]),
			PublicKey:      base64.RawURLEncoding.EncodeToString(publicKey),
		}},
	}
	if err := set.add("notifications-consumer-manifest.json", manifest); err != nil {
		return err
	}
	if err := set.add("notifications-consumer-signing-seed", []byte(
		base64.RawURLEncoding.EncodeToString(seed),
	)); err != nil {
		return err
	}
	if err := addJSON(set, "notifications-consumer-keyring.json", consumerRing); err != nil {
		return err
	}
	kek, err := randomBase64(32)
	if err != nil {
		return err
	}
	set.trackExpiry(expiresAt)
	return addJSON(set, "notifications-kek-keyring.json", notificationsKEKKeyringDocument{
		Version: 1, Current: notificationsKEKDocument{Revision: 1, Key: kek},
		Retained: []notificationsKEKDocument{},
	})
}
