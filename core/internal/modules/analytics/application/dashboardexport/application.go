package dashboardexport

import (
	"fmt"
	"reflect"
)

type Application struct {
	binder     StoreBinder
	authorizer Authorizer
	ids        IDSource
	clock      Clock
	scheduler  Scheduler
	content    ContentRuntime
}

func NewApplication(
	binder StoreBinder, authorizer Authorizer, ids IDSource, clock Clock,
	scheduler Scheduler, content ContentRuntime,
) (*Application, error) {
	for _, dependency := range []any{binder, authorizer, ids, clock, scheduler, content} {
		if nilPort(dependency) {
			return nil, fmt.Errorf("analytics dashboard export dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer, ids: ids,
		clock: clock, scheduler: scheduler, content: content}, nil
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
