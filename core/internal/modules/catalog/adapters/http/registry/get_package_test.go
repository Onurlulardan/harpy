package registryhttp

import "testing"

func TestGetPackage(t *testing.T) {
	if err := exercisePackage("get_package", "catalog.package.get",
		func(handler *Handler) Route { return handler.GetPackageRoute() }); err != nil {
		t.Fatal(err)
	}
}
