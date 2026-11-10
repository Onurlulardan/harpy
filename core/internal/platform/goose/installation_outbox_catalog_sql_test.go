package goose

import (
	"os"
	"strings"
	"testing"
)

func TestInstallationOutboxUsesTheProducerCatalogIdentity(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00803_platform_async_installation_catalog_persist.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	up, _, found := strings.Cut(string(body), "-- +goose Down")
	if !found {
		t.Fatal("down section is missing")
	}
	for _, fragment := range []string{
		"installation.activated",
		"'iam'",
		"4a44e2f3f9e2669280d713512812010d50bf9f077f4ee1a0709e80c3633dcb5e",
		"delivery_state = 'delivered'",
		"kind IN ('installation.activation.recorded', 'iam.initial_admin_setup.deliver')",
		"'succeeded'",
	} {
		if !strings.Contains(up, fragment) {
			t.Fatalf("missing %s", fragment)
		}
	}
	if strings.Contains(up, "'installation.activation.recorded', 1, 'tenancy_onboarding_operation'") {
		t.Fatal("activation event is still written outside the producer catalog")
	}
}

func TestInstallationOutboxWriterStaysImmutable(t *testing.T) {
	body, err := os.ReadFile(
		"sql/platform/00798_platform_async_installation_persist.sql",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "'installation.activation.recorded'") {
		t.Fatal("applied installation activation migration changed")
	}
}
