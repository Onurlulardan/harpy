package resolved

import "bytes"

func cloneValue(value Value) Value {
	result := value
	result.json = bytes.Clone(value.json)
	return result
}

func cloneValues(values []Value) []Value {
	result := make([]Value, len(values))
	for index, value := range values {
		result[index] = cloneValue(value)
	}
	return result
}
