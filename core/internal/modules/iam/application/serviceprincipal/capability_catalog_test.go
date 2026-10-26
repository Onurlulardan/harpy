package serviceprincipal

import "testing"

func TestCapabilityCatalogRequiresRegisteredExactScope(t *testing.T) {
	key, err := ParseCapabilityKey("forms.record.update")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCapabilityCatalog([]CapabilityDefinition{{
		Key: key, Owner: "forms", ResourceKind: "record", ResourceScope: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !catalog.Allows(key, ResourceScope{
		Kind: ScopeResource, Owner: "forms", ResourceKind: "record",
		ResourceID: [16]byte{1},
	}) {
		t.Fatal("registered resource capability was denied")
	}
	if catalog.Allows(key, ResourceScope{
		Kind: ScopeTenant, Owner: "forms", ResourceKind: "record",
	}) {
		t.Fatal("undeclared tenant scope was allowed")
	}
}
