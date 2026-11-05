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
		"installation-delivery prepare", "run --rm core-operator", "status -ne 5",
		"task local:provision-tenant", "run --rm core-operator", "publish-local-authority",
		":local:inventory:verify", "installation-delivery deliver", "installation-delivery verify",
	})
	provision := taskBlock(t, tasks, "provision-tenant", "services")
	assertOrdered(t, provision, []string{
		"show-provision-request", "run --rm --no-deps",
		"HARPY_PROVISION_TENANT_ID=${tenant_id}", "tenant-provisioning",
	})
	api := runtime[:strings.Index(runtime, "\n  tenant-provisioning:\n")]
	if !strings.HasSuffix(strings.TrimSpace(provision), "tenant-provisioning") ||
		strings.Contains(tasks, "docker.sock") || strings.Contains(runtime, "docker.sock") ||
		strings.Count(runtime, "local-installation-state:/run/installation") != 1 ||
		strings.Count(runtime, "env_file: ./state/secrets/tenant-provisioning.env") != 1 ||
		!strings.Contains(runtime, "restart: \"no\"") ||
		!strings.Contains(api, "local-installation-state:/run/installation") ||
		strings.Contains(api, "tenant-provisioning.env") {
		t.Fatal("installation host wiring broke tenant provisioning")
	}
}
