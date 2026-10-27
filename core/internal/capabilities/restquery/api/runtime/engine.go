package runtime

import (
	"fmt"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/executor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/authorization"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type Engine struct {
	catalogs   CatalogPort
	authorizer authorization.Authorizer
	executor   executor.Executor
}

func New(
	owner string,
	catalogs CatalogPort,
	policy PolicyPort,
	ownerPort OwnerPort,
	keys CursorKeys,
) (*Engine, error) {
	if nilInterface(catalogs) {
		return nil, fmt.Errorf("REST Query runtime dependencies are invalid")
	}
	return newEngine(owner, catalogs, policy, ownerPort, keys)
}

func newEngine(
	owner string,
	catalogs CatalogPort,
	policy PolicyPort,
	ownerPort OwnerPort,
	keys CursorKeys,
) (*Engine, error) {
	if owner == "" || nilInterface(policy) || nilInterface(ownerPort) {
		return nil, fmt.Errorf("REST Query runtime dependencies are invalid")
	}
	policies, err := authorization.NewPolicyRegistry([]authorization.PolicyRegistration{{
		Owner: owner, Port: policy,
	}})
	if err != nil {
		return nil, err
	}
	authorizer, err := authorization.NewAuthorizer(policies)
	if err != nil {
		return nil, err
	}
	registration, err := execution.NewRegistration(owner, ownerPort)
	if err != nil {
		return nil, err
	}
	owners, err := execution.NewRegistry([]execution.Registration{registration})
	if err != nil {
		return nil, err
	}
	keyring, err := keys.keyring()
	if err != nil {
		return nil, err
	}
	codec, err := cursor.NewCodec(keyring)
	if err != nil {
		return nil, err
	}
	database, err := executor.New(owners, codec, planner.HardPolicy())
	if err != nil {
		return nil, err
	}
	return &Engine{catalogs: catalogs, authorizer: authorizer, executor: database}, nil
}

func nilInterface(value any) bool {
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
