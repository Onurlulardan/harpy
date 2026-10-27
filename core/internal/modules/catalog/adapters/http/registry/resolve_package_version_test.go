package registryhttp

import "testing"

func TestResolvePackageVersion(t *testing.T) {
	if err := exercisePackageVersion("get_package_version", "catalog.package_version.get",
		func(handler *Handler) Route { return handler.ResolvePackageVersionRoute() }); err != nil {
		t.Fatal(err)
	}
}
