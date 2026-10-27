package api

import "testing"

func TestCatalogSetRejectsDuplicateIdentityOrHashConflict(t *testing.T) {
	_, rootCatalog, targetCatalog, rootReference, _ := mustCatalogSetFixture(t)
	if _, err := NewCatalogSet(
		rootReference, []FieldCatalog{rootCatalog, rootCatalog, targetCatalog},
	); err == nil {
		t.Fatal("duplicate exact catalog accepted")
	}
	conflictingReference := mustCatalogReferenceFixture(
		t, "forms", rootResourceID, rootVersionID, "4",
	)
	conflictingCatalog := mustFieldCatalogFixture(t, catalogFixtureInput(
		t, conflictingReference, "conflict_records",
		targetStableField, targetSearchField, nil,
	))
	if _, err := NewCatalogSet(
		rootReference, []FieldCatalog{rootCatalog, targetCatalog, conflictingCatalog},
	); err == nil {
		t.Fatal("one catalog identity accepted conflicting hashes")
	}
}

func TestCatalogSetRejectsRootOrRelationHashMismatch(t *testing.T) {
	_, rootCatalog, targetCatalog, rootReference, _ := mustCatalogSetFixture(t)
	wrongRoot := mustCatalogReferenceFixture(
		t, "forms", rootResourceID, rootVersionID, "4",
	)
	if _, err := NewCatalogSet(
		wrongRoot, []FieldCatalog{rootCatalog, targetCatalog},
	); err == nil {
		t.Fatal("hash-mismatched root reference accepted")
	}
	wrongTarget := mustCatalogReferenceFixture(
		t, "crm", targetResourceID, targetVersionID, "4",
	)
	relation, _ := NewRelationDefinition(
		rootRelationID, "customer", RelationCardinalityToOne,
		mustRelationStorage(t, rootRelationID), wrongTarget, true, 1,
	)
	unresolvedRoot := mustFieldCatalogFixture(t, catalogFixtureInput(
		t, rootReference, "unresolved_records", rootStableField, rootSearchField,
		[]RelationDefinition{relation},
	))
	if _, err := NewCatalogSet(
		rootReference, []FieldCatalog{unresolvedRoot, targetCatalog},
	); err == nil {
		t.Fatal("unresolved relation target accepted")
	}
}

func TestCatalogSetZeroAndTamperFailClosed(t *testing.T) {
	valid, _, _, _, _ := mustCatalogSetFixture(t)
	unvalidated := valid
	unvalidated.validated = false
	missingCatalogs := valid
	missingCatalogs.catalogs = nil
	invalidRoot := valid
	invalidRoot.root = CatalogReference{}
	invalidCatalog := valid
	invalidCatalog.catalogs = cloneFieldCatalogs(valid.catalogs)
	invalidCatalog.catalogs[0].fields[0].publicName = "_included"
	for _, set := range []CatalogSet{
		{}, unvalidated, missingCatalogs, invalidRoot, invalidCatalog,
	} {
		if set.Validate() == nil {
			t.Fatal("invalid catalog set validated")
		}
		if _, err := set.Catalogs(); err == nil {
			t.Fatal("invalid catalog set accessor succeeded")
		}
	}
}
