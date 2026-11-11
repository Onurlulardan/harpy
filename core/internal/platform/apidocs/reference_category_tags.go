package apidocs

import (
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/contracts"
)

func referenceTag(ref string) (string, bool, error) {
	file, pointer, found := strings.Cut(ref, "#")
	if !found || file == "" || strings.Contains(file, "..") {
		return "", false, nil
	}
	raw, err := contracts.Files.ReadFile("http/" + file)
	if err != nil {
		return "", false, err
	}
	var document any
	if err = json.Unmarshal(raw, &document); err != nil {
		return "", false, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		if part == "" {
			continue
		}
		object, ok := document.(map[string]any)
		if !ok {
			return "", false, nil
		}
		document, ok = object[part]
		if !ok {
			return "", false, nil
		}
	}
	tag, found := methodTag(document)
	return tag, found, nil
}

func methodTag(value any) (string, bool) {
	object, ok := value.(map[string]any)
	if !ok {
		return "", false
	}
	for _, method := range []string{"get", "post", "put", "patch", "delete", "head"} {
		operation, found := object[method].(map[string]any)
		if !found {
			continue
		}
		tags, _ := operation["tags"].([]any)
		if len(tags) != 1 {
			continue
		}
		tag, ok := tags[0].(string)
		if ok && tag != "" {
			return tag, true
		}
	}
	return "", false
}
