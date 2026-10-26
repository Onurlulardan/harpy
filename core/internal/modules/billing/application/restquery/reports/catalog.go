package reports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type catalogEntry struct {
	spec      resourceSpec
	reference api.CatalogReference
	set       api.CatalogSet
}

type Catalog struct {
	entries map[string]catalogEntry
	valid   bool
}

func NewCatalog() (*Catalog, error) {
	entries := make(map[string]catalogEntry, len(resources()))
	for _, spec := range resources() {
		entry, err := buildCatalogEntry(spec)
		if err != nil {
			return nil, err
		}
		entries[spec.Name] = entry
	}
	catalog := &Catalog{entries: entries, valid: true}
	if catalog.Validate() != nil {
		return nil, ErrContract
	}
	return catalog, nil
}

func (catalog *Catalog) Validate() error {
	if catalog == nil || !catalog.valid || len(catalog.entries) != len(resources()) {
		return ErrContract
	}
	for _, entry := range catalog.entries {
		if entry.reference.Validate() != nil || entry.set.Validate() != nil ||
			entry.spec.Name == "" {
			return ErrContract
		}
	}
	return nil
}

func (catalog *Catalog) Catalog(
	ctx context.Context,
	scope transaction.GlobalScope,
	resource string,
) (api.CatalogSet, error) {
	entry, found := catalog.entries[resource]
	if catalog.Validate() != nil || !found || ctx == nil || ctx.Err() != nil ||
		transaction.NetworkAllowed(ctx) || !transaction.ContextOwnsGlobalScope(ctx, scope) ||
		transaction.RequireGlobalOwner(scope, Owner) != nil {
		return api.CatalogSet{}, ErrContract
	}
	return entry.set, nil
}
