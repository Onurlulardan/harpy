package lifecyclehttp

import "testing"

func TestInstall(t *testing.T) {
	if err := exerciseInstallation("install", "catalog.installation.install",
		func(handler *Handler) Route { return handler.InstallRoute() }); err != nil {
		t.Fatal(err)
	}
}
