package config

import (
	"encoding/json"
	"time"
	"unicode/utf8"
)

type iamTimedHMACKey struct {
	key         iamHMACKey
	activatedAt time.Time
	issueUntil  time.Time
	verifyUntil time.Time
	verifyNull  bool
}

func loadIAMHMACKeyring(
	values environmentValues,
	environmentKey string,
	purpose iamHMACPurpose,
	now time.Time,
) (iamHMACKeyring, error) {
	secret, err := loadSecret(values, environmentKey, true, secretFileLimit)
	if err != nil {
		return iamHMACKeyring{}, err
	}
	keyring, ok := parseIAMHMACKeyring(secret, purpose, now)
	if !ok {
		return iamHMACKeyring{}, invalidIAMKeyring(environmentKey, "iam-hmac-keyring.v1")
	}
	return keyring, nil
}

func parseIAMHMACKeyring(
	secret Secret,
	expected iamHMACPurpose,
	now time.Time,
) (iamHMACKeyring, bool) {
	document := []byte(secret.Value())
	if !utf8.Valid(document) || !json.Valid(document) ||
		rejectDuplicateJSONMembers(document) != nil ||
		!expected.valid() || now.IsZero() {
		return iamHMACKeyring{}, false
	}
	var root map[string]json.RawMessage
	if json.Unmarshal(document, &root) != nil ||
		!hasExactMembers(root, "version", "purpose", "active", "retained") {
		return iamHMACKeyring{}, false
	}
	var version int
	var purpose iamHMACPurpose
	if json.Unmarshal(root["version"], &version) != nil || version != 1 ||
		json.Unmarshal(root["purpose"], &purpose) != nil || purpose != expected {
		return iamHMACKeyring{}, false
	}
	active, ok := decodeIAMActiveHMACKey(root["active"])
	if !ok || active.issueUntil.Sub(active.activatedAt) != iamHMACRotation ||
		now.Before(active.activatedAt) || !now.Before(active.issueUntil) {
		return iamHMACKeyring{}, false
	}
	var retainedDocuments []json.RawMessage
	if json.Unmarshal(root["retained"], &retainedDocuments) != nil || len(retainedDocuments) > 64 {
		return iamHMACKeyring{}, false
	}
	retained, ok := decodeIAMRetainedHMACKeys(retainedDocuments, purpose, active, now)
	if !ok {
		return iamHMACKeyring{}, false
	}
	return iamHMACKeyring{
		purpose: purpose, active: active.key, retained: retained, valid: true,
	}, true
}

func decodeIAMActiveHMACKey(raw json.RawMessage) (iamTimedHMACKey, bool) {
	object, ok := decodeIAMObject(raw, "kid", "key", "activated_at", "issue_until")
	if !ok {
		return iamTimedHMACKey{}, false
	}
	kid, materialText, activatedText, issueText, ok := decodeIAMStrings(
		object, "kid", "key", "activated_at", "issue_until",
	)
	material, materialOK := decodeIAMKeyMaterial(materialText)
	activatedAt, activatedOK := parseIAMTimestamp(activatedText)
	issueUntil, issueOK := parseIAMTimestamp(issueText)
	if !ok || !validIAMHMACKeyID(kid) || !materialOK || iamAllZero(material[:]) ||
		!activatedOK || !issueOK {
		return iamTimedHMACKey{}, false
	}
	return iamTimedHMACKey{
		key:         iamHMACKey{kid: kid, material: material},
		activatedAt: activatedAt, issueUntil: issueUntil,
	}, true
}
