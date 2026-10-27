package discount

import (
	"errors"
	"fmt"

	queryruntime "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/runtime"
)

var ErrContract = errors.New("billing discount REST Query contract failed")

type CursorKeys = queryruntime.CursorKeys
type GlobalOwnerPort = queryruntime.GlobalOwnerPort

type Service struct{ engine *queryruntime.GlobalEngine }

func NewService(ownerPort GlobalOwnerPort, keys CursorKeys) (*Service, error) {
	if ownerPort == nil {
		return nil, fmt.Errorf("billing discount REST Query owner is required")
	}
	catalog, err := NewCatalog()
	if err != nil {
		return nil, err
	}
	policy, err := NewPolicy(catalog)
	if err != nil {
		return nil, err
	}
	engine, err := queryruntime.NewGlobal(Owner, catalog, policy, ownerPort, keys)
	if err != nil {
		return nil, err
	}
	return &Service{engine: engine}, nil
}
