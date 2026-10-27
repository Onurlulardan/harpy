package publication

import (
	"fmt"
	"reflect"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
)

type Application struct {
	binder       StoreBinder
	authorizer   Authorizer
	entitlements billingapi.MeteredResourceMutationDecider
	resources    ResourceVerifier
	ids          IDSource
	random       RandomSource
	clock        Clock
	effects      Effects
}

func NewApplication(
	binder StoreBinder, authorizer Authorizer,
	entitlements billingapi.MeteredResourceMutationDecider, resources ResourceVerifier,
	ids IDSource, random RandomSource, clock Clock, effects Effects,
) (*Application, error) {
	for _, value := range []any{
		binder, authorizer, entitlements, resources, ids, random, clock, effects,
	} {
		if nilPort(value) {
			return nil, fmt.Errorf("analytics publication dependencies are invalid")
		}
	}
	return &Application{binder: binder, authorizer: authorizer,
		entitlements: entitlements, resources: resources,
		ids: ids, random: random, clock: clock, effects: effects}, nil
}

type PublicApplication struct {
	directory PublicDirectory
	scopes    PublicScopeCoordinator
	executor  PublicExecutor
	limiter   PublicLimiter
	clock     Clock
}

func NewPublicApplication(
	directory PublicDirectory, scopes PublicScopeCoordinator,
	executor PublicExecutor, limiter PublicLimiter, clock Clock,
) (*PublicApplication, error) {
	for _, value := range []any{directory, scopes, executor, limiter, clock} {
		if nilPort(value) {
			return nil, fmt.Errorf("analytics public dependencies are invalid")
		}
	}
	return &PublicApplication{directory: directory, scopes: scopes,
		executor: executor, limiter: limiter, clock: clock}, nil
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
