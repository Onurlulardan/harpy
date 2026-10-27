package platformiam

import "testing"

func TestCapabilityCatalogIsExactAndUnique(t *testing.T) {
	if len(CapabilityCatalog) != 49 {
		t.Fatalf("capability count=%d, want 49", len(CapabilityCatalog))
	}
	seen := make(map[string]struct{}, len(CapabilityCatalog))
	for _, capability := range CapabilityCatalog {
		if _, duplicate := seen[capability]; duplicate || !CapabilityKnown(capability) {
			t.Fatalf("invalid capability catalog entry %q", capability)
		}
		seen[capability] = struct{}{}
	}
	for _, forbidden := range []string{"platform.*", "environment_all", "tenant_all", "catalog.read"} {
		if CapabilityKnown(forbidden) {
			t.Fatalf("forbidden capability %q accepted", forbidden)
		}
	}
}
