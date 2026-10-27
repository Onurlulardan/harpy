package runtime

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/executor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/authorization"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type GlobalEngine struct {
	catalogs   GlobalCatalogPort
	authorizer authorization.GlobalAuthorizer
	executor   executor.GlobalExecutor
	codec      cursor.Codec
}

func NewGlobal(
	owner string,
	catalogs GlobalCatalogPort,
	policy GlobalPolicyPort,
	ownerPort GlobalOwnerPort,
	keys CursorKeys,
) (*GlobalEngine, error) {
	if owner == "" || nilInterface(catalogs) || nilInterface(policy) ||
		nilInterface(ownerPort) {
		return nil, fmt.Errorf("REST Query global runtime dependencies are invalid")
	}
	authorizer, err := authorization.NewGlobalAuthorizer(owner, policy)
	if err != nil {
		return nil, err
	}
	registration, err := execution.NewGlobalRegistration(owner, ownerPort)
	if err != nil {
		return nil, err
	}
	owners, err := execution.NewGlobalRegistry([]execution.GlobalRegistration{registration})
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
	database, err := executor.NewGlobal(owners, codec, planner.HardPolicy())
	if err != nil {
		return nil, err
	}
	return &GlobalEngine{
		catalogs: catalogs, authorizer: authorizer,
		executor: database, codec: codec,
	}, nil
}
