package billingv1adoptionapp

import (
	"encoding/json"
	"fmt"
)

var featureFields = []string{"ai_chat", "workflows", "api_access"}

func decodeFeatures(document []byte) (map[string]bool, []MissingField, error) {
	allowed := make(map[string]struct{}, len(featureFields))
	for _, field := range featureFields {
		allowed[field] = struct{}{}
	}
	values, err := decodeObject(document, allowed)
	if err != nil {
		return nil, nil, ErrFeatures
	}
	result := make(map[string]bool, len(featureFields))
	missing := make([]MissingField, 0, len(featureFields))
	for _, field := range featureFields {
		raw, exists := values[field]
		if !exists {
			result[field] = false
			missing = append(missing, MissingField{"features", field})
			continue
		}
		var value bool
		if string(raw) == "null" || json.Unmarshal(raw, &value) != nil {
			return nil, nil, fmt.Errorf("%w: %s", ErrFeatures, field)
		}
		result[field] = value
	}
	return result, missing, nil
}
