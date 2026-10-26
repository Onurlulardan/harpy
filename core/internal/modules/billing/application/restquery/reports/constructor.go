package reports

import (
	"fmt"

	queryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
)

type QueryRequest = queryapi.QueryRequest

type CursorKeys = queryruntime.CursorKeys
type GlobalOwnerPort = queryruntime.GlobalOwnerPort

func NewCursorKeys(
	currentID string,
	current [32]byte,
	previousID string,
	previous [32]byte,
) (CursorKeys, error) {
	return queryruntime.NewCursorKeys(currentID, current, previousID, previous)
}

type Service struct{ engine *queryruntime.GlobalEngine }

func NewService(ownerPort GlobalOwnerPort, keys CursorKeys) (*Service, error) {
	if ownerPort == nil {
		return nil, fmt.Errorf("billing report REST Query owner is required")
	}
	catalogs, err := NewCatalog()
	if err != nil {
		return nil, err
	}
	policy, err := NewPolicy(catalogs)
	if err != nil {
		return nil, err
	}
	engine, err := queryruntime.NewGlobal(Owner, catalogs, policy, ownerPort, keys)
	if err != nil {
		return nil, err
	}
	return &Service{engine: engine}, nil
}
