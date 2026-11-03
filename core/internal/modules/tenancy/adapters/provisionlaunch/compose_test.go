package provisionlaunch

import (
	"strings"
	"testing"
)

func TestComposeArgsUseFixedTenantProvisioningService(t *testing.T) {
	args := strings.Join(composeArgs("00112233-4455-6677-8899-aabbccddeeff"), " ")
	if !strings.Contains(args, "run --rm --no-deps") ||
		!strings.Contains(args, "-f ops/local/compose.runtime.yml") ||
		!strings.HasSuffix(args, "tenant-provisioning") ||
		strings.Contains(args, "docker.sock") {
		t.Fatal(args)
	}
	if strings.Count(args, "HARPY_PROVISION_TENANT_ID=00112233-4455-6677-8899-aabbccddeeff") != 1 {
		t.Fatal(args)
	}
}
