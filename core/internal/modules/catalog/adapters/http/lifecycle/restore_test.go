package lifecyclehttp

import "testing"

func TestRestore(t *testing.T) {
	if err := exerciseInstallation("restore", "catalog.installation.restore",
		func(handler *Handler) Route { return handler.RestoreRoute() }); err != nil {
		t.Fatal(err)
	}
}
