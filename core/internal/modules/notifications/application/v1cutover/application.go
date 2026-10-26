package v1cutover

import (
	"context"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

type Application struct{ binder Binder }

func NewApplication(binder Binder) (*Application, error) {
	if cutoverNil(binder) {
		return nil, ErrInvalid
	}
	return &Application{binder: binder}, nil
}

func (application *Application) store(ctx context.Context, scope api.Scope) (Store, error) {
	if application == nil || ctx == nil || scope == nil {
		return nil, ErrInvalid
	}
	store, err := application.binder.BindCutover(ctx, scope)
	if err != nil || cutoverNil(store) {
		return nil, ErrUnavailable
	}
	return store, nil
}

func cutoverNil(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return (seen.Kind() == reflect.Pointer || seen.Kind() == reflect.Interface ||
		seen.Kind() == reflect.Func || seen.Kind() == reflect.Map || seen.Kind() == reflect.Slice) && seen.IsNil()
}
