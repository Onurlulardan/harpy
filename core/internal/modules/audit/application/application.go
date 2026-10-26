package application

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type Application struct {
	binder     StoreBinder
	authorizer Authorizer
	ids        IDSource
	clock      Clock
}

func NewApplication(
	binder StoreBinder, authorizer Authorizer, ids IDSource, clock Clock,
) (*Application, error) {
	for _, dependency := range []any{binder, authorizer, ids, clock} {
		if nilPort(dependency) {
			return nil, fmt.Errorf("audit application dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer, ids: ids, clock: clock}, nil
}

func (application *Application) newID() (uuid.UUID, error) {
	if application == nil || nilPort(application.ids) {
		return uuid.Nil, ErrUnavailable
	}
	id, err := application.ids.NewUUIDv7()
	if err != nil || id == uuid.Nil || id.Version() != 7 {
		return uuid.Nil, ErrUnavailable
	}
	return id, nil
}

func (application *Application) now() (time.Time, error) {
	if application == nil || nilPort(application.clock) {
		return time.Time{}, ErrUnavailable
	}
	now := application.clock.Now().UTC().Truncate(time.Microsecond)
	if now.IsZero() || now.Nanosecond()%1000 != 0 {
		return time.Time{}, ErrUnavailable
	}
	return now, nil
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
