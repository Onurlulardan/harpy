package cursor

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func tryBuildContext(fixture contextFixture) (Context, error) {
	hash, err := api.NewCatalogHash(fixture.hash)
	if err != nil {
		return Context{}, err
	}
	reference, err := api.NewCatalogReference(
		fixture.owner, fixture.resource, fixture.version, hash,
	)
	if err != nil {
		return Context{}, err
	}
	return NewContext(
		fixture.tenant, fixture.subject, fixture.auth, reference, fixture.query,
		fixture.mandatory, fixture.pageSize, fixture.sorts,
	)
}
