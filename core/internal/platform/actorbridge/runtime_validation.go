package actorbridge

import "reflect"

func interfaceIsNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func runtimeFault(err error, fallback FaultCode) error {
	if err == nil {
		return nil
	}
	for _, code := range []FaultCode{
		FaultInvalid, FaultUnavailable, FaultAuthority, FaultRetired,
		FaultReplay, FaultExpired, FaultUnauthorized, FaultUnreconciled,
	} {
		if FaultIs(err, code) {
			return err
		}
	}
	return NewFault(fallback)
}
