package local

import (
	"strings"
	"testing"
)

func TestInstallationDeliverProvisionsPendingTenant(t *testing.T) {
	tasks := readContractText(t, "../../.taskfiles/local.yml")
	runtime := readContractText(t, "compose.runtime.yml")
	deliver := taskBlock(t, tasks, "installation:deliver", "provision-tenant")
	assertOrdered(t, deliver, []string{
		"run --rm installation-delivery prepare-initial-credential",
		"go run ./ops/local/cmd/provision-request wait-operator",
		"run --rm core-operator",
		"run --rm installation-delivery prepare\\\"",
		"publish-local-authority", ":local:inventory:verify",
		"installation-delivery deliver", "installation-delivery verify",
	})
	if strings.Contains(deliver, "tr ") || strings.Contains(deliver, "case ") ||
		strings.Contains(deliver, "status -ne 5") {
		t.Fatal(deliver)
	}
	provision := taskBlock(t, tasks, "provision-tenant", "services")
	if strings.TrimSpace(provision) == "" ||
		!strings.Contains(provision, "go run ./ops/local/cmd/provision-request") {
		t.Fatal(provision)
	}
	for _, forbidden := range []string{
		"tr ", "printf", "case ", "sed ", "awk ", "set -eu", "bash",
		"show-provision-request", "docker compose",
	} {
		if strings.Contains(provision, forbidden) {
			t.Fatalf("provision-tenant contains %q", forbidden)
		}
	}
	command := readContractText(t, "cmd/provision-request/compose_args.go")
	assertOrdered(t, command, []string{
		"run", "-T", "--rm", "--no-deps", "show-provision-request",
		"HARPY_PROVISION_TENANT_ID=", "tenant-provisioning",
	})
	api := runtime[:strings.Index(runtime, "\n  tenant-provisioning:\n")]
	if strings.Contains(command, "docker.sock") ||
		strings.Contains(tasks, "docker.sock") || strings.Contains(runtime, "docker.sock") ||
		strings.Count(runtime, "local-installation-state:/run/installation") != 1 ||
		strings.Count(runtime, "env_file: ./state/secrets/tenant-provisioning.env") != 1 ||
		!strings.Contains(runtime, "restart: \"no\"") ||
		!strings.Contains(api, "local-installation-state:/run/installation") ||
		strings.Contains(api, "tenant-provisioning.env") {
		t.Fatal("installation host wiring broke tenant provisioning")
	}
}
