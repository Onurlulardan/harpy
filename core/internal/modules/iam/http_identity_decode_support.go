package iam

import (
	"encoding/json"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
	"github.com/google/uuid"
)

const identityHTTPBodyLimit int64 = 64 * 1024

func decodeIdentityDocument[T any](
	w http.ResponseWriter,
	r *http.Request,
) (T, error) {
	var document T
	if err := httpinput.DecodeJSON(w, r, identityHTTPBodyLimit, &document); err != nil {
		return document, err
	}
	return document, nil
}

func decodeIdentityProfileDocument[T any](
	w http.ResponseWriter,
	r *http.Request,
	knownFields []string,
) (T, json.RawMessage, error) {
	var document T
	var members map[string]json.RawMessage
	if err := httpinput.DecodeJSON(w, r, identityHTTPBodyLimit, &members); err != nil {
		return document, nil, err
	}
	known := make(map[string]struct{}, len(knownFields))
	for _, field := range knownFields {
		known[field] = struct{}{}
	}
	core := make(map[string]json.RawMessage, len(members))
	profile := make(map[string]json.RawMessage)
	for key, value := range members {
		if key == "custom_attributes" {
			if err := mergeIdentityProfileInput(profile, value); err != nil {
				return document, nil, err
			}
			continue
		}
		if _, found := known[key]; found {
			core[key] = value
			continue
		}
		if !validIdentityProfileKey(key) || identityProfileInputReserved(key) {
			return document, nil, httpinput.ErrInvalid
		}
		if _, collision := profile[key]; collision {
			return document, nil, httpinput.ErrInvalid
		}
		profile[key] = value
	}
	coreJSON, coreErr := json.Marshal(core)
	profileJSON, profileErr := json.Marshal(profile)
	if coreErr != nil || profileErr != nil || json.Unmarshal(coreJSON, &document) != nil {
		return document, nil, httpinput.ErrInvalid
	}
	if len(profile) == 0 {
		profileJSON = nil
	}
	return document, profileJSON, nil
}

func mergeIdentityProfileInput(
	target map[string]json.RawMessage, raw json.RawMessage,
) error {
	values, err := identityProfileInputMap(raw)
	if err != nil {
		return httpinput.ErrInvalid
	}
	for key, value := range values {
		if _, collision := target[key]; collision {
			return httpinput.ErrInvalid
		}
		target[key] = value
	}
	return nil
}

func identityProfileInputReserved(key string) bool {
	switch key {
	case "id", "revision", "created_at", "updated_at", "password_hash",
		"email_normalized", "token_version", "authorization_version",
		"identifier_id", "profile_schema_version_id", "identity_kind",
		"portal_identifier", "user_ids", "owner_ids", "owner_id":
		return true
	default:
		return false
	}
}

func decodeOptionalUUID(raw *string) (*uuid.UUID, bool) {
	if raw == nil {
		return nil, true
	}
	parsed, err := uuid.Parse(*raw)
	if err != nil || parsed == uuid.Nil || parsed.String() != *raw {
		return nil, false
	}
	return &parsed, true
}

func decodeUUID(raw string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(raw)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == raw
}

func decodeUUIDList(values []string) ([]uuid.UUID, bool) {
	result := make([]uuid.UUID, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for index, value := range values {
		parsed, valid := decodeUUID(value)
		if _, duplicate := seen[parsed]; !valid || duplicate {
			return nil, false
		}
		seen[parsed] = struct{}{}
		result[index] = parsed
	}
	return result, len(result) <= 100
}

func identityCustomAttributes(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	return append(json.RawMessage(nil), raw...)
}
