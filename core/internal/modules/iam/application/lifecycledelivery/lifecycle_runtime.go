package lifecycledelivery

import "reflect"

type Runtime struct{ store Store }

func NewRuntime(store Store) (*Runtime, error) {
	if nilStore(store) {
		return nil, ErrInvalid
	}
	return &Runtime{store: store}, nil
}

func nilStore(store Store) bool {
	if store == nil {
		return true
	}
	value := reflect.ValueOf(store)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
