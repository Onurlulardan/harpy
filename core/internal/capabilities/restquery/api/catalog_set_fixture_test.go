package api

import "testing"

func mustCatalogSetFixture(
	t *testing.T,
) (CatalogSet, FieldCatalog, FieldCatalog, CatalogReference, CatalogReference) {
	t.Helper()
	rootReference := mustCatalogReferenceFixture(
		t, "forms", rootResourceID, rootVersionID, "1",
	)
	targetReference := mustCatalogReferenceFixture(
		t, "crm", targetResourceID, targetVersionID, "2",
	)
	targetCatalog := mustFieldCatalogFixture(t, catalogFixtureInput(
		t, targetReference, "target_records", targetStableField, targetSearchField, nil,
	))
	relation, err := NewRelationDefinition(
		rootRelationID, "customer", RelationCardinalityToOne,
		mustRelationStorage(t, rootRelationID), targetReference, true, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	rootCatalog := mustFieldCatalogFixture(t, catalogFixtureInput(
		t, rootReference, "root_records", rootStableField, rootSearchField,
		[]RelationDefinition{relation},
	))
	set, err := NewCatalogSet(
		rootReference, []FieldCatalog{rootCatalog, targetCatalog},
	)
	if err != nil {
		t.Fatal(err)
	}
	return set, rootCatalog, targetCatalog, rootReference, targetReference
}
