package lifecyclehttp

import "testing"

func TestUpgrade(t *testing.T) {
	if err := exerciseInstallation("upgrade", "catalog.installation.upgrade",
		func(handler *Handler) Route { return handler.UpgradeRoute() }); err != nil {
		t.Fatal(err)
	}
}
