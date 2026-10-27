package jsonscan

import (
	"encoding/json"
	"strconv"
)

func scanValue(
	decoder *json.Decoder, pointer string, depth int, state *scanState,
) error {
	token, err := decoder.Token()
	if err != nil {
		return ErrInvalidDocument
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	if delimiter != '{' && delimiter != '[' {
		return ErrInvalidDocument
	}
	if depth >= state.maximumDepth {
		return ErrNestingLimit
	}
	if delimiter == '{' {
		return scanObject(decoder, pointer, depth, state)
	}
	return scanArray(decoder, pointer, depth, state)
}

func scanObject(
	decoder *json.Decoder, pointer string, depth int, state *scanState,
) error {
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		name, stringName := token.(string)
		if err != nil || !stringName {
			return ErrInvalidDocument
		}
		memberPointer := appendPointer(pointer, name)
		if _, exists := seen[name]; exists {
			state.duplicates[memberPointer] = struct{}{}
		} else {
			seen[name] = struct{}{}
		}
		if err := scanValue(decoder, memberPointer, depth+1, state); err != nil {
			return err
		}
	}
	return consumeDelimiter(decoder, '}')
}

func scanArray(
	decoder *json.Decoder, pointer string, depth int, state *scanState,
) error {
	index := 0
	for decoder.More() {
		itemPointer := appendPointer(pointer, strconv.Itoa(index))
		if err := scanValue(decoder, itemPointer, depth+1, state); err != nil {
			return err
		}
		index++
	}
	return consumeDelimiter(decoder, ']')
}

func consumeDelimiter(decoder *json.Decoder, expected json.Delim) error {
	token, err := decoder.Token()
	delimiter, valid := token.(json.Delim)
	if err != nil || !valid || delimiter != expected {
		return ErrInvalidDocument
	}
	return nil
}
