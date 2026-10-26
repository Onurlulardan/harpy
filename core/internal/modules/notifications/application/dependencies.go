package application

import (
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type Application struct {
	binder     ports.Binder
	recipients ports.RecipientResolver
	authorizer ports.ProducerAuthorizer
	clock      ports.Clock
	ids        ports.IDSource
}

func NewApplication(binder ports.Binder, recipients ports.RecipientResolver,
	authorizer ports.ProducerAuthorizer, clock ports.Clock, ids ports.IDSource,
) (*Application, error) {
	if nilDependency(binder) || nilDependency(recipients) || nilDependency(authorizer) ||
		nilDependency(clock) || nilDependency(ids) {
		return nil, ErrConfiguration
	}
	return &Application{binder: binder, recipients: recipients,
		authorizer: authorizer, clock: clock, ids: ids}, nil
}

func nilDependency(value any) bool {
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
