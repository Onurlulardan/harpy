package api

import (
	"os"
	"strings"
	"testing"
)

func TestIAMActorBridgeConsumesStartupKeyAfterHTTPComposition(t *testing.T) {
	assembly, err := os.ReadFile("assemble_api_runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(assembly)
	billingForms := strings.Index(source, "buildAPIBillingAndForms(")
	domains := strings.Index(source, "buildAPIDomainRuntimes(")
	actorBridge := strings.Index(source, "iamRuntime.buildActorBridge(settings)")
	if billingForms < 0 || domains < 0 || actorBridge < 0 {
		t.Fatal("IAM security composition steps are missing")
	}
	if actorBridge < billingForms || actorBridge < domains {
		t.Fatal("actor bridge must consume and wipe its startup key after every IAM config consumer")
	}
	portal, err := os.ReadFile("build_api_billing_forms.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(portal), "iamRuntime.buildPortalUserHTTP(") {
		t.Fatal("portal HTTP must consume IAM config before actor bridge")
	}
	provisioning, err := os.ReadFile("build_api_domain_runtimes.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(provisioning), "buildCustomerProvisioningRuntime(") {
		t.Fatal("customer provisioning must consume IAM config before actor bridge")
	}
}
