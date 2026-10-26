package compiler

import (
	"strconv"

	"github.com/google/uuid"
)

type binder struct {
	arguments []any
}

func newBinder() *binder { return &binder{arguments: []any{}} }

func (binder *binder) bind(value any) (string, error) {
	if binder == nil || !validArgument(value) {
		return "", ErrContract
	}
	binder.arguments = append(binder.arguments, cloneArgument(value))
	return "$" + strconv.Itoa(len(binder.arguments)), nil
}

func validArguments(values []any) bool {
	for _, value := range values {
		if !validArgument(value) {
			return false
		}
	}
	return true
}

func validArgument(value any) bool {
	switch typed := value.(type) {
	case string, bool, int, int64:
		return true
	case []string:
		return typed != nil
	case []uuid.UUID:
		if typed == nil {
			return false
		}
		for _, identifier := range typed {
			if identifier == uuid.Nil || identifier.Variant() != uuid.RFC4122 {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func cloneArguments(values []any) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = cloneArgument(value)
	}
	return result
}

func cloneArgument(value any) any {
	if values, ok := value.([]string); ok {
		return append([]string{}, values...)
	}
	if values, ok := value.([]uuid.UUID); ok {
		return append([]uuid.UUID{}, values...)
	}
	return value
}
