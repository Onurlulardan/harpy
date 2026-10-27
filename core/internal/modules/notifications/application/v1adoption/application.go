package v1adoption

import (
	"context"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

type Application struct{ binder Binder }

func NewApplication(binder Binder) (*Application, error) {
	if nilAdoptionPort(binder) {
		return nil, ErrInvalid
	}
	return &Application{binder: binder}, nil
}

func (application *Application) Step(ctx context.Context, scope api.Scope,
	plan Plan,
) (StepResult, error) {
	tenantID, tenantScoped := tenantIdentity(scope)
	if application == nil || ctx == nil || scope == nil || plan.Validate() != nil ||
		!tenantScoped || tenantID != plan.TenantID {
		return StepResult{}, ErrInvalid
	}
	store, err := application.binder.BindAdoption(ctx, scope)
	if err != nil || store == nil {
		return StepResult{}, ErrUnavailable
	}
	runner, err := NewRunner(store)
	if err != nil {
		return StepResult{}, err
	}
	return runner.Execute(ctx, plan)
}

func nilAdoptionPort(value any) bool {
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
