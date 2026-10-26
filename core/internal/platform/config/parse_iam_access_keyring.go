package config

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"time"
	"unicode/utf8"
)

type iamRetiredAccessKey struct {
	kid          string
	publicKey    ed25519.PublicKey
	signUntil    time.Time
	publishUntil time.Time
}

func loadIAMAccessKeyring(
	values environmentValues,
	now time.Time,
) (IAMAccessTokenKeyring, error) {
	secret, err := loadSecret(values, iamAccessKeyringEnvironmentKey, true, secretFileLimit)
	if err != nil {
		return IAMAccessTokenKeyring{}, err
	}
	keyring, ok := parseIAMAccessKeyring(secret, now)
	if !ok {
		return IAMAccessTokenKeyring{}, invalidIAMKeyring(
			iamAccessKeyringEnvironmentKey, "iam-access-token-keyring.v1",
		)
	}
	return keyring, nil
}

func parseIAMAccessKeyring(secret Secret, now time.Time) (IAMAccessTokenKeyring, bool) {
	document := []byte(secret.Value())
	if !utf8.Valid(document) || !json.Valid(document) ||
		rejectDuplicateJSONMembers(document) != nil || now.IsZero() {
		return IAMAccessTokenKeyring{}, false
	}
	var root map[string]json.RawMessage
	if json.Unmarshal(document, &root) != nil ||
		!hasExactMembers(root, "version", "active", "retired_public") {
		return IAMAccessTokenKeyring{}, false
	}
	var version int
	if json.Unmarshal(root["version"], &version) != nil || version != 1 {
		return IAMAccessTokenKeyring{}, false
	}
	active, activatedAt, signUntil, ok := decodeIAMActiveAccessKey(root["active"])
	if !ok || signUntil.Sub(activatedAt) != iamAccessSigningInterval ||
		now.Before(activatedAt) || !now.Before(signUntil) {
		return IAMAccessTokenKeyring{}, false
	}
	var retiredDocuments []json.RawMessage
	if json.Unmarshal(root["retired_public"], &retiredDocuments) != nil || len(retiredDocuments) > 8 {
		return IAMAccessTokenKeyring{}, false
	}
	retired, ok := decodeIAMRetiredAccessKeys(retiredDocuments, active, activatedAt, now)
	if !ok {
		return IAMAccessTokenKeyring{}, false
	}
	return IAMAccessTokenKeyring{
		active: active, retired: retired, valid: true,
	}, true
}

func decodeIAMActiveAccessKey(raw json.RawMessage) (
	iamAccessSigningKey,
	time.Time,
	time.Time,
	bool,
) {
	object, ok := decodeIAMObject(raw, "kid", "seed", "activated_at", "sign_until")
	if !ok {
		return iamAccessSigningKey{}, time.Time{}, time.Time{}, false
	}
	kid, seedText, activatedText, signText, ok := decodeIAMStrings(
		object, "kid", "seed", "activated_at", "sign_until",
	)
	seed, seedOK := decodeIAMKeyMaterial(seedText)
	activatedAt, activatedOK := parseIAMTimestamp(activatedText)
	signUntil, signOK := parseIAMTimestamp(signText)
	if !ok || !validIAMAccessKeyID(kid) || !seedOK || iamAllZero(seed[:]) ||
		!activatedOK || !signOK {
		return iamAccessSigningKey{}, time.Time{}, time.Time{}, false
	}
	return iamAccessSigningKey{kid: kid, seed: seed}, activatedAt, signUntil, true
}

func decodeIAMRetiredAccessKeys(
	documents []json.RawMessage,
	active iamAccessSigningKey,
	activeStart time.Time,
	now time.Time,
) (map[string]ed25519.PublicKey, bool) {
	keys := make(map[string]ed25519.PublicKey, len(documents))
	var previousSignUntil time.Time
	var previousPublicKey ed25519.PublicKey
	for index, document := range documents {
		retired, ok := decodeIAMRetiredAccessKey(document)
		if !ok || !now.Before(retired.publishUntil) ||
			retired.publishUntil.Before(retired.signUntil.Add(24*time.Hour)) ||
			(index > 0 && retired.signUntil.Sub(previousSignUntil) !=
				iamAccessSigningInterval) {
			return nil, false
		}
		if retired.kid == active.kid || iamAccessPublicMatchesSeed(retired.publicKey, active.seed) {
			return nil, false
		}
		if index > 0 && bytes.Equal(previousPublicKey, retired.publicKey) {
			return nil, false
		}
		for _, existing := range keys {
			if bytes.Equal(existing, retired.publicKey) {
				return nil, false
			}
		}
		if _, duplicate := keys[retired.kid]; duplicate {
			return nil, false
		}
		keys[retired.kid] = append(ed25519.PublicKey(nil), retired.publicKey...)
		previousSignUntil = retired.signUntil
		previousPublicKey = retired.publicKey
	}
	if len(documents) > 0 && !previousSignUntil.Equal(activeStart) {
		return nil, false
	}
	return keys, true
}
