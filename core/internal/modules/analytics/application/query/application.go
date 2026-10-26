package query

import (
	"fmt"
	"reflect"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
)

type Application struct {
	binder       StoreBinder
	authorizer   Authorizer
	entitlements billingapi.MeteredResourceMutationDecider
	dependencies DependencyResolver
	contracts    ContractPort
	cache        Cache
	ids          IDSource
	clock        Clock
	effects      Effects
	references   FormsReferenceCoordinator
}

func NewApplication(
	binder StoreBinder, authorizer Authorizer,
	entitlements billingapi.MeteredResourceMutationDecider, dependencies DependencyResolver,
	contracts ContractPort, cache Cache, ids IDSource, clock Clock, effects Effects,
	references FormsReferenceCoordinator,
) (*Application, error) {
	for _, value := range []any{
		binder, authorizer, entitlements, dependencies, contracts, cache, ids, clock,
		effects, references,
	} {
		if nilPort(value) {
			return nil, fmt.Errorf("analytics query dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer, entitlements: entitlements,
		dependencies: dependencies, contracts: contracts, cache: cache,
		ids: ids, clock: clock, effects: effects, references: references}, nil
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
