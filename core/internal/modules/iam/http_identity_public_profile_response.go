package iam

import (
	"encoding/json"
	"fmt"
)

func identityFlattenPublicProfile(record any) (map[string]json.RawMessage, error) {
	document, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal IAM identity record: %w", err)
	}
	var result map[string]json.RawMessage
	if json.Unmarshal(document, &result) != nil || result == nil {
		return nil, ErrIdentityHTTPUnavailable
	}
	raw, found := result["custom_attributes"]
	delete(result, "custom_attributes")
	if !found {
		return result, nil
	}
	var profile map[string]json.RawMessage
	if json.Unmarshal(raw, &profile) != nil || len(profile) > 100 {
		return nil, ErrIdentityHTTPUnavailable
	}
	for key, value := range profile {
		if !validIdentityProfileKey(key) {
			return nil, ErrIdentityHTTPUnavailable
		}
		if _, collision := result[key]; collision {
			return nil, ErrIdentityHTTPUnavailable
		}
		result[key] = value
	}
	return result, nil
}
