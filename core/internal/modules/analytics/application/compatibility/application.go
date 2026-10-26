package compatibility

import (
	"fmt"
	"reflect"
)

type Application struct {
	adoption  AdoptionStore
	legacy    PublicLegacyStore
	lifecycle LegacyStore
	authority RetirementAuthority
	clock     Clock
}

func NewApplication(
	adoption AdoptionStore, legacy LegacyStore, authority RetirementAuthority, clock Clock,
) (*Application, error) {
	for _, value := range []any{adoption, legacy, authority, clock} {
		if nilPort(value) {
			return nil, fmt.Errorf("analytics compatibility dependencies are invalid")
		}
	}
	return &Application{adoption: adoption, legacy: legacy, lifecycle: legacy,
		authority: authority, clock: clock}, nil
}

func NewLegacyApplication(
	legacy PublicLegacyStore, clock Clock,
) (*Application, error) {
	if nilPort(legacy) || nilPort(clock) {
		return nil, fmt.Errorf("analytics legacy compatibility dependencies are invalid")
	}
	return &Application{legacy: legacy, clock: clock}, nil
}

func nilPort(value any) bool {
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
