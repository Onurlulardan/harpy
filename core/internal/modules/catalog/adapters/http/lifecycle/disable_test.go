package lifecyclehttp

import "testing"

func TestDisable(t *testing.T) {
	if err := exerciseInstallation("disable", "catalog.installation.disable",
		func(handler *Handler) Route { return handler.DisableRoute() }); err != nil {
		t.Fatal(err)
	}
}
