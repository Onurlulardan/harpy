package execution

import (
	"reflect"

	restexecution "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type DriverPort interface {
	Execute(transaction.TenantScope, restexecution.Program) error
}

type Executor struct{ port DriverPort }

func NewExecutor(port DriverPort) (*Executor, error) {
	if portIsNil(port) {
		return nil, ErrContract
	}
	return &Executor{port: port}, nil
}

func portIsNil(port DriverPort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
