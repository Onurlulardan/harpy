package compiler

import "reflect"

func includeConsumerIsNil(consumer any) bool {
	if consumer == nil {
		return true
	}
	value := reflect.ValueOf(consumer)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
