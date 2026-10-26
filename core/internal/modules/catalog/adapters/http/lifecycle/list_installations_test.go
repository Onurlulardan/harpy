package lifecyclehttp

import "testing"

func TestListInstallations(t *testing.T) {
	if err := exerciseInstallation("list_installations", "catalog.installation.list",
		func(handler *Handler) Route { return handler.ListInstallationsRoute() }); err != nil {
		t.Fatal(err)
	}
}
