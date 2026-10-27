package dashboard

import (
	"fmt"
	"reflect"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
)

type Application struct {
	binder       StoreBinder
	authorizer   Authorizer
	entitlements billingapi.MeteredResourceMutationDecider
	queries      QueryVersionVerifier
	ids          IDSource
	clock        Clock
	effects      Effects
}

func NewApplication(
	binder StoreBinder, authorizer Authorizer,
	entitlements billingapi.MeteredResourceMutationDecider, queries QueryVersionVerifier,
	ids IDSource, clock Clock, effects Effects,
) (*Application, error) {
	for _, value := range []any{
		binder, authorizer, entitlements, queries, ids, clock, effects,
	} {
		if nilPort(value) {
			return nil, fmt.Errorf("analytics dashboard dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer,
		entitlements: entitlements, queries: queries,
		ids: ids, clock: clock, effects: effects}, nil
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
