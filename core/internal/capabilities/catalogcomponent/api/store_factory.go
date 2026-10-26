package catalogcomponent

import (
	"context"
	"fmt"
	"regexp"

	catalogpostgres "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogcomponent/adapters/postgres"
)

var ownerPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type Queries struct {
	Read     string
	Apply    string
	Blockers string
	Allows   string
}

type Row = catalogpostgres.Row
type BoundStore = catalogpostgres.Store
type QueryRow func(context.Context, string, ...any) Row
type ScopeGuard func(context.Context) bool

type StoreFactory struct{ adapter *catalogpostgres.Factory }

func NewStoreFactory(owner string, queries Queries) (*StoreFactory, error) {
	if !ownerPattern.MatchString(owner) {
		return nil, fmt.Errorf("catalog component store owner is invalid")
	}
	adapter, err := catalogpostgres.NewFactory(owner, catalogpostgres.Queries{
		Read: queries.Read, Apply: queries.Apply,
		Blockers: queries.Blockers, Allows: queries.Allows,
	})
	if err != nil {
		return nil, err
	}
	return &StoreFactory{adapter: adapter}, nil
}

func (factory *StoreFactory) Build(
	generation int64, queryRow QueryRow, allowed ScopeGuard,
) (*BoundStore, error) {
	if factory == nil || factory.adapter == nil || queryRow == nil || allowed == nil {
		return nil, fmt.Errorf("catalog component store factory is unavailable")
	}
	return factory.adapter.Build(generation,
		catalogpostgres.QueryRow(queryRow), catalogpostgres.ScopeGuard(allowed))
}
