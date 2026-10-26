package config

import "encoding/json"

func decodeIAMObject(raw json.RawMessage, names ...string) (map[string]json.RawMessage, bool) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || !hasExactMembers(object, names...) {
		return nil, false
	}
	return object, true
}

func decodeIAMStrings(
	object map[string]json.RawMessage,
	names ...string,
) (string, string, string, string, bool) {
	if len(names) != 4 {
		return "", "", "", "", false
	}
	values := make([]string, 4)
	for index, name := range names {
		if json.Unmarshal(object[name], &values[index]) != nil {
			return "", "", "", "", false
		}
	}
	return values[0], values[1], values[2], values[3], true
}

func invalidIAMKeyring(key, contract string) error {
	return configError(key, "must satisfy "+contract)
}
