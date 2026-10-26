package cryptography

import (
	"fmt"
	"log/slog"
	"reflect"
)

const authenticationMaterialRedacted = "[REDACTED IAM AUTHENTICATION]"

type secureMaterial struct {
	value []byte
	valid bool
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func newSecureMaterial(value []byte, minimum, maximum int) (secureMaterial, error) {
	if len(value) < minimum || len(value) > maximum {
		return secureMaterial{}, fmt.Errorf("authentication material is invalid")
	}
	return secureMaterial{value: append([]byte(nil), value...), valid: true}, nil
}

func (material *secureMaterial) copy() ([]byte, error) {
	if material == nil || !material.valid || len(material.value) == 0 {
		return nil, fmt.Errorf("authentication material is unavailable")
	}
	return append([]byte(nil), material.value...), nil
}

func (material *secureMaterial) Destroy() {
	if material == nil {
		return
	}
	clear(material.value)
	material.value, material.valid = nil, false
}

func (secureMaterial) String() string   { return authenticationMaterialRedacted }
func (secureMaterial) GoString() string { return authenticationMaterialRedacted }
func (secureMaterial) LogValue() slog.Value {
	return slog.StringValue(authenticationMaterialRedacted)
}
func (secureMaterial) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED IAM AUTHENTICATION]"`), nil
}
func (secureMaterial) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(authenticationMaterialRedacted))
}
