package sessioncapability

import "testing"

func TestAdmissionRequiresModeBeforeExactCapability(t *testing.T) {
	sets := Sets{
		Mode: ModeOnPremise, Purpose: PurposeOnPremisePrimary,
		PlatformCapabilities: []string{
			"platform.billing.catalog.read", "platform.instance.read",
			"platform.license.read",
		},
		TenantCapabilities: []string{
			"billing.entitlements.read", "catalog.read",
		},
	}
	if !sets.AllowsPlatform("platform.instance.read") ||
		!sets.AllowsPlatform("platform.license.read") ||
		sets.AllowsPlatform("platform.billing.catalog.read") {
		t.Fatal("platform mode guard did not precede exact capability membership")
	}
	if !sets.AllowsTenant("catalog.read") ||
		!sets.AllowsTenant("billing.entitlements.read") ||
		sets.AllowsTenant("billing.subscription.read") {
		t.Fatal("tenant mode guard did not precede exact capability membership")
	}
}

func TestAdmissionHasNoBooleanIdentityBypass(t *testing.T) {
	sets := Sets{
		Mode: ModeSaaS, Purpose: PurposeSaaSCustomer,
		PlatformCapabilities: []string{}, TenantCapabilities: []string{"catalog.read"},
	}
	for _, forged := range []string{
		"is_super_admin", "superadmin@harpy.io", "x-platform-admin",
		"route.platform.billing.catalog.read", "platform.*", "tenant_all",
	} {
		if sets.AllowsPlatform(forged) || sets.AllowsTenant(forged) {
			t.Fatalf("forged boolean identity bypass was accepted: %s", forged)
		}
	}
}

func TestCatalogContainsEveryPhase14TenantCapability(t *testing.T) {
	if len(TenantReleaseCatalog) != 17 {
		t.Fatalf("unexpected release catalog size: %d", len(TenantReleaseCatalog))
	}
	for _, capability := range TenantReleaseCatalog {
		if !TenantReleaseCapabilityKnown(capability) {
			t.Fatalf("release capability is not exact: %s", capability)
		}
	}
}
