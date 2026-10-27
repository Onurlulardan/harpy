package api

import "testing"

func TestCatalogSetExactRootAndRelationResolution(t *testing.T) {
	set, rootCatalog, targetCatalog, rootReference, targetReference := mustCatalogSetFixture(t)
	root, rootErr := set.Root()
	catalogs, catalogsErr := set.Catalogs()
	actualRoot, rootCatalogErr := set.RootCatalog()
	actualTarget, found, targetErr := set.Catalog(targetReference)
	if rootErr != nil || catalogsErr != nil || rootCatalogErr != nil || targetErr != nil {
		t.Fatal("valid catalog set accessor failed")
	}
	if root != rootReference || len(catalogs) != 2 || !found ||
		actualRoot.reference != rootCatalog.reference ||
		actualTarget.reference != targetCatalog.reference {
		t.Fatal("catalog set did not resolve exact references")
	}
	wrongHash := mustCatalogReferenceFixture(
		t, "crm", targetResourceID, targetVersionID, "3",
	)
	if _, found, err := set.Catalog(wrongHash); err != nil || found {
		t.Fatal("hash-mismatched reference did not return a clean miss")
	}
	if _, _, err := set.Catalog(CatalogReference{}); err == nil {
		t.Fatal("invalid catalog lookup reference accepted")
	}
}

func TestCatalogSetDeepImmutability(t *testing.T) {
	_, rootCatalog, targetCatalog, rootReference, _ := mustCatalogSetFixture(t)
	inputs := []FieldCatalog{rootCatalog, targetCatalog}
	set, err := NewCatalogSet(rootReference, inputs)
	if err != nil {
		t.Fatal(err)
	}
	inputs[0].fields[0].publicName = "mutated"
	if set.Validate() != nil {
		t.Fatal("catalog set retained constructor input storage")
	}
	actual, _ := set.Catalogs()
	actual[0].fields[0].publicName = "mutated"
	actual[0].defaultSelect = nil
	if set.Validate() != nil {
		t.Fatal("catalog set accessor exposed internal storage")
	}
}
