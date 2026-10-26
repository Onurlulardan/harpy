package registryhttp

import "testing"

func TestCompilePackageVersions(t *testing.T) {
	if err := exercisePackageVersion("list_package_versions", "catalog.package_version.list",
		func(handler *Handler) Route { return handler.CompilePackageVersionsRoute() }); err != nil {
		t.Fatal(err)
	}
}
