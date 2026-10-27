package published

import (
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolvedTemplateValues(values []resolved.Value, list bool) (any, error) {
	if !list && len(values) != 1 {
		return nil, fmt.Errorf("published query template arity is invalid")
	}
	result := make([]any, len(values))
	for index, value := range values {
		encoded, err := encodeResolvedTemplateValue(value)
		if err != nil {
			return nil, err
		}
		result[index] = encoded
	}
	if list {
		return result, nil
	}
	return result[0], nil
}

func encodeResolvedTemplateValue(value resolved.Value) (any, error) {
	if _, err := value.Kind(); err != nil {
		return nil, err
	}
	if boolean, found, accessorErr := value.Boolean(); accessorErr == nil && found {
		return boolean, nil
	}
	if document, found, accessorErr := value.JSON(); accessorErr == nil && found {
		return json.RawMessage(document), nil
	}
	scalar, found, err := value.Scalar()
	if err != nil || !found {
		return nil, fmt.Errorf("published query template value is invalid")
	}
	return scalar, nil
}

func applyTemplatePage(root any, page PageOverride) (any, error) {
	document, ok := root.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("published query template root is invalid")
	}
	pageDocument, _ := document["page"].(map[string]any)
	if pageDocument == nil {
		pageDocument = map[string]any{}
	}
	if size, present, err := page.Size(); err != nil {
		return nil, err
	} else if present {
		pageDocument["size"] = size
	}
	if after, present, err := page.After(); err != nil {
		return nil, err
	} else if present {
		pageDocument["after"] = after
	}
	document["page"] = pageDocument
	return document, nil
}
