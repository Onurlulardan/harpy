package artifact

import (
	"encoding/json"
	"fmt"
	"math/big"
)

var maximumCanonicalInteger = big.NewInt(9_007_199_254_740_991)

func encodeCanonicalValue(target []byte, value any) ([]byte, error) {
	switch current := value.(type) {
	case nil:
		return append(target, "null"...), nil
	case bool:
		if current {
			return append(target, "true"...), nil
		}
		return append(target, "false"...), nil
	case string:
		encoded, _ := json.Marshal(current)
		return append(target, encoded...), nil
	case json.Number:
		integer, valid := new(big.Int).SetString(current.String(), 10)
		if !valid || new(big.Int).Abs(integer).Cmp(maximumCanonicalInteger) > 0 {
			return nil, fmt.Errorf("non-interoperable JSON number")
		}
		return append(target, current.String()...), nil
	case []any:
		return encodeCanonicalArray(target, current)
	case map[string]any:
		return encodeCanonicalObject(target, current)
	default:
		return nil, fmt.Errorf("unsupported JSON value")
	}
}

func encodeCanonicalArray(target []byte, value []any) ([]byte, error) {
	target = append(target, '[')
	for index, item := range value {
		if index != 0 {
			target = append(target, ',')
		}
		var err error
		target, err = encodeCanonicalValue(target, item)
		if err != nil {
			return nil, err
		}
	}
	return append(target, ']'), nil
}

func encodeCanonicalObject(target []byte, value map[string]any) ([]byte, error) {
	target = append(target, '{')
	for index, key := range sortedKeys(value) {
		if index != 0 {
			target = append(target, ',')
		}
		encodedKey, _ := json.Marshal(key)
		target = append(target, encodedKey...)
		target = append(target, ':')
		var err error
		target, err = encodeCanonicalValue(target, value[key])
		if err != nil {
			return nil, err
		}
	}
	return append(target, '}'), nil
}
