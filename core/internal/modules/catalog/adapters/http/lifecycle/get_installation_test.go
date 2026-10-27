package lifecyclehttp

import "testing"

func TestGetInstallation(t *testing.T) {
	if err := exerciseInstallation("get_installation", "catalog.installation.get",
		func(handler *Handler) Route { return handler.GetInstallationRoute() }); err != nil {
		t.Fatal(err)
	}
}
