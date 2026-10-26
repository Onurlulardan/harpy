package resolver

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type resolvedValueKey struct {
	kind    api.LogicalType
	state   uint8
	scalar  string
	boolean bool
	json    string
}

func deduplicateValues(values []resolved.Value) ([]resolved.Value, error) {
	result := make([]resolved.Value, 0, len(values))
	seen := make(map[resolvedValueKey]struct{}, len(values))
	for _, value := range values {
		key, err := resolvedKey(value)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func resolvedKey(value resolved.Value) (resolvedValueKey, error) {
	kind, err := value.Kind()
	if err != nil {
		return resolvedValueKey{}, fmt.Errorf("resolved value key is invalid")
	}
	if scalar, found, accessorErr := value.Scalar(); accessorErr != nil {
		return resolvedValueKey{}, fmt.Errorf("resolved value key is invalid")
	} else if found {
		return resolvedValueKey{kind: kind, state: 1, scalar: scalar}, nil
	}
	if boolean, found, accessorErr := value.Boolean(); accessorErr != nil {
		return resolvedValueKey{}, fmt.Errorf("resolved value key is invalid")
	} else if found {
		return resolvedValueKey{kind: kind, state: 2, boolean: boolean}, nil
	}
	if document, found, accessorErr := value.JSON(); accessorErr != nil || !found {
		return resolvedValueKey{}, fmt.Errorf("resolved value key is invalid")
	} else {
		return resolvedValueKey{kind: kind, state: 3, json: string(document)}, nil
	}
}
