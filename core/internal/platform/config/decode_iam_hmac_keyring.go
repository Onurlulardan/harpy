package config

import (
	"bytes"
	"encoding/json"
	"time"
)

func decodeIAMRetainedHMACKeys(
	documents []json.RawMessage,
	purpose iamHMACPurpose,
	active iamTimedHMACKey,
	now time.Time,
) (map[string][32]byte, bool) {
	keys := make(map[string][32]byte, len(documents))
	var previous iamTimedHMACKey
	for index, document := range documents {
		key, ok := decodeIAMRetainedHMACKey(document)
		if !ok || key.issueUntil.Sub(key.activatedAt) != iamHMACRotation ||
			!validIAMRetention(key, purpose, now) {
			return nil, false
		}
		if index > 0 && !previous.issueUntil.Equal(key.activatedAt) {
			return nil, false
		}
		if key.key.kid == active.key.kid || bytes.Equal(key.key.material[:], active.key.material[:]) {
			return nil, false
		}
		for retainedID, retainedMaterial := range keys {
			if retainedID == key.key.kid || bytes.Equal(retainedMaterial[:], key.key.material[:]) {
				return nil, false
			}
		}
		keys[key.key.kid] = key.key.material
		previous = key
	}
	if len(documents) > 0 && !previous.issueUntil.Equal(active.activatedAt) {
		return nil, false
	}
	return keys, true
}

func decodeIAMRetainedHMACKey(raw json.RawMessage) (iamTimedHMACKey, bool) {
	object, ok := decodeIAMObject(
		raw, "kid", "key", "activated_at", "issue_until", "verify_until",
	)
	if !ok {
		return iamTimedHMACKey{}, false
	}
	kid, materialText, activatedText, issueText, ok := decodeIAMStrings(
		object, "kid", "key", "activated_at", "issue_until",
	)
	material, materialOK := decodeIAMKeyMaterial(materialText)
	activatedAt, activatedOK := parseIAMTimestamp(activatedText)
	issueUntil, issueOK := parseIAMTimestamp(issueText)
	key := iamTimedHMACKey{
		key:         iamHMACKey{kid: kid, material: material},
		activatedAt: activatedAt, issueUntil: issueUntil,
	}
	if bytes.Equal(bytes.TrimSpace(object["verify_until"]), []byte("null")) {
		key.verifyNull = true
	} else {
		var verifyText string
		if json.Unmarshal(object["verify_until"], &verifyText) != nil {
			return iamTimedHMACKey{}, false
		}
		key.verifyUntil, ok = parseIAMTimestamp(verifyText)
		if !ok {
			return iamTimedHMACKey{}, false
		}
	}
	if !ok || !validIAMHMACKeyID(kid) || !materialOK || iamAllZero(material[:]) ||
		!activatedOK || !issueOK {
		return iamTimedHMACKey{}, false
	}
	return key, true
}

func validIAMRetention(key iamTimedHMACKey, purpose iamHMACPurpose, now time.Time) bool {
	if purpose.receiptGated() {
		return key.verifyNull
	}
	return !key.verifyNull && now.Before(key.verifyUntil) &&
		!key.verifyUntil.Before(key.issueUntil.Add(purpose.minimumRetention()))
}
