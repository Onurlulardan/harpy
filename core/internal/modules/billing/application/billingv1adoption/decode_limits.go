package billingv1adoptionapp

import (
	"encoding/json"
	"fmt"
)

var limitFields = []string{
	"submissions_per_month", "max_users", "max_forms", "max_storage_mb",
}

func decodeLimits(document []byte) (map[string]int64, []MissingField, error) {
	allowed := make(map[string]struct{}, len(limitFields))
	for _, field := range limitFields {
		allowed[field] = struct{}{}
	}
	values, err := decodeObject(document, allowed)
	if err != nil {
		return nil, nil, ErrLimits
	}
	result := make(map[string]int64, len(limitFields))
	missing := make([]MissingField, 0, len(limitFields))
	for _, field := range limitFields {
		raw, exists := values[field]
		if !exists {
			result[field] = 0
			missing = append(missing, MissingField{"limits", field})
			continue
		}
		var value int64
		if string(raw) == "null" || json.Unmarshal(raw, &value) != nil || value < -1 {
			return nil, nil, fmt.Errorf("%w: %s", ErrLimits, field)
		}
		result[field] = value
	}
	return result, missing, nil
}
