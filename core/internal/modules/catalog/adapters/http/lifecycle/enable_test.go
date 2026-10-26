package lifecyclehttp

import "testing"

func TestEnable(t *testing.T) {
	if err := exerciseInstallation("enable", "catalog.installation.enable",
		func(handler *Handler) Route { return handler.EnableRoute() }); err != nil {
		t.Fatal(err)
	}
}
