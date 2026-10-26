package tenancy

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
)

var (
	errCustomerLifecyclePrecondition = errors.New("customer lifecycle precondition is required")
	errCustomerLifecycleStatsMoved   = errors.New("customer lifecycle stats moved")
	errCustomerLifecycleStatusMoved  = errors.New("customer lifecycle status mutation moved")
	errCustomerLifecycleSwitchGone   = errors.New("customer lifecycle tenant switching is retired")
)

type CustomerLifecycleApplication interface {
	List(context.Context, lifecycle.ListQuery) (lifecycle.Page, error)
	Get(context.Context, lifecycle.GetQuery) (lifecycle.Tenant, error)
	Update(context.Context, lifecycle.MutationCommand) (lifecycle.Tenant, error)
	Suspend(context.Context, lifecycle.MutationCommand) (lifecycle.Tenant, error)
	Reactivate(context.Context, lifecycle.MutationCommand) (lifecycle.Tenant, error)
	RequestOffboard(context.Context, lifecycle.AsyncCommand) (lifecycle.Operation, error)
	RequestPurge(context.Context, lifecycle.AsyncCommand) (lifecycle.Operation, error)
	ReadOperation(context.Context, lifecycle.OperationQuery) (lifecycle.Operation, error)
}

type CustomerLifecycleHTTP struct {
	application CustomerLifecycleApplication
	report      func(error)
}

func NewCustomerLifecycleHTTP(application CustomerLifecycleApplication,
	report func(error),
) (*CustomerLifecycleHTTP, error) {
	if customerLifecycleNil(application) || report == nil {
		return nil, fmt.Errorf("customer lifecycle HTTP dependencies are required")
	}
	return &CustomerLifecycleHTTP{application: application, report: report}, nil
}

func customerLifecycleNil(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}
