package registryhttp

import "testing"

func TestListPackages(t *testing.T) {
	if err := exercisePackage("list_packages", "catalog.package.list",
		func(handler *Handler) Route { return handler.ListPackagesRoute() }); err != nil {
		t.Fatal(err)
	}
}
