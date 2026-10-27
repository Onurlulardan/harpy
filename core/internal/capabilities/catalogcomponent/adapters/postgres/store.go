package catalogcomponentpostgres

import (
	"context"
	"fmt"
	"regexp"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var ownerPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type Row interface{ Scan(...any) error }
type QueryRow func(context.Context, string, ...any) Row
type ScopeGuard func(context.Context) bool

type Factory struct {
	owner   string
	queries Queries
}

func NewFactory(owner string, queries Queries) (*Factory, error) {
	if !ownerPattern.MatchString(owner) || queries.validate() != nil {
		return nil, fmt.Errorf("catalog component store dependencies are invalid")
	}
	return &Factory{owner: owner, queries: queries}, nil
}

func (factory *Factory) Build(
	generation int64, queryRow QueryRow, allowed ScopeGuard,
) (*Store, error) {
	if factory == nil || generation < 1 || queryRow == nil || allowed == nil ||
		!ownerPattern.MatchString(factory.owner) || factory.queries.validate() != nil {
		return nil, fmt.Errorf("catalog component store binding is invalid")
	}
	return &Store{owner: factory.owner, generation: generation,
		queries: factory.queries, queryRow: queryRow, allowed: allowed}, nil
}

type Store struct {
	owner      string
	generation int64
	queries    Queries
	queryRow   QueryRow
	allowed    ScopeGuard
}

func (storage *Store) usable(ctx context.Context) bool {
	return storage != nil && storage.owner != "" && storage.generation > 0 &&
		storage.queries.validate() == nil && storage.queryRow != nil &&
		storage.allowed != nil && ctx != nil && ctx.Err() == nil && storage.allowed(ctx)
}
