package billing

import (
	"strings"
	"testing"
)

func TestSaaSFreeBaseMigrationHasNoTrialOrSubscriptionLifecycle(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var seed, assignment string
	for _, document := range documents {
		path := document.Path()
		switch {
		case strings.Contains(path, "20260722110001_seed_billing_saas_free_base_v1") &&
			strings.Contains(path, "/fresh/"):
			seed = string(document.Content())
		case strings.Contains(path, "20260722110003_create_billing_installation_assign_v2") &&
			strings.Contains(path, "/fresh/"):
			assignment = string(document.Content())
		}
	}
	for _, required := range []string{
		"'saas_free_base_v1'", "true,false,true,0,0,1,13",
		"'monthly',0,'USD'", "'subscription',false,'trial',false",
	} {
		if !strings.Contains(seed, required) {
			t.Fatalf("SaaS free base seed lacks %q", required)
		}
	}
	for _, required := range []string{
		"p_binding_purpose='saas_customer'",
		"platform.billing_saas_free_base_publications",
		"INSERT INTO platform.billing_tenant_base_offerings",
	} {
		if !strings.Contains(assignment, required) {
			t.Fatalf("SaaS free base assignment lacks %q", required)
		}
	}
	for _, forbidden := range []string{
		"INSERT INTO platform.billing_subscriptions",
		"INSERT INTO platform.billing_subscription_trials",
		"billing_subscription_create",
	} {
		if strings.Contains(assignment, forbidden) {
			t.Fatalf("SaaS free base assignment creates forbidden lifecycle %q", forbidden)
		}
	}
}
