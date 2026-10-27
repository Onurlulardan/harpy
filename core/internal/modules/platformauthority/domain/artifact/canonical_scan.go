package artifact

import (
	"encoding/json"
	"fmt"
)

const maximumCanonicalDepth = 64

func scanCanonicalValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > maximumCanonicalDepth {
		return nil, fmt.Errorf("JSON nesting exceeds its limit")
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, composite := token.(json.Delim); composite {
		switch delimiter {
		case '{':
			return scanCanonicalObject(decoder, depth+1)
		case '[':
			return scanCanonicalArray(decoder, depth+1)
		default:
			return nil, fmt.Errorf("unexpected JSON delimiter")
		}
	}
	return token, nil
}

func scanCanonicalObject(decoder *json.Decoder, depth int) (map[string]any, error) {
	value := map[string]any{}
	for decoder.More() {
		keyToken, err := decoder.Token()
		key, ok := keyToken.(string)
		if err != nil || !ok {
			return nil, fmt.Errorf("invalid JSON object key")
		}
		if _, duplicate := value[key]; duplicate {
			return nil, fmt.Errorf("duplicate JSON object key")
		}
		item, err := scanCanonicalValue(decoder, depth)
		if err != nil {
			return nil, err
		}
		value[key] = item
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, fmt.Errorf("unterminated JSON object")
	}
	return value, nil
}

func scanCanonicalArray(decoder *json.Decoder, depth int) ([]any, error) {
	value := []any{}
	for decoder.More() {
		item, err := scanCanonicalValue(decoder, depth)
		if err != nil {
			return nil, err
		}
		value = append(value, item)
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim(']') {
		return nil, fmt.Errorf("unterminated JSON array")
	}
	return value, nil
}
