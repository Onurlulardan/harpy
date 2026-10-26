package postgres

import "reflect"

var boundExecutorType = reflect.TypeOf(BoundExecutor{})

func rawBoundExecutorIdentity(value any) bool {
	valueType := reflect.TypeOf(value)
	return valueType == boundExecutorType ||
		valueType == reflect.PointerTo(boundExecutorType)
}
