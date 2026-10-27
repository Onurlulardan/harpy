package foundation

import (
	"reflect"
	"time"
)

func NilPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

func ValidInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() != nil
}

func ValidIdempotency(value string) bool {
	return len(value) >= 8 && len(value) <= 256
}
