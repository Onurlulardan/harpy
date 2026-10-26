package api

import "testing"

func TestCatalogSetAllowsExactLogicalStorageReuseAcrossVersions(t *testing.T) {
	_, root, target, rootReference, _ := mustCatalogSetFixture(t)
	alternate := root.clone()
	alternate.reference = mustCatalogReferenceFixture(
		t, "forms", rootResourceID,
		"90000000-0000-4000-8000-000000000001", "5",
	)
	if _, err := NewCatalogSet(
		rootReference, []FieldCatalog{root, alternate, target},
	); err != nil {
		t.Fatal("exact logical storage reuse across versions rejected")
	}
}

func TestCatalogSetRejectsCrossLogicalResourceTableReuse(t *testing.T) {
	_, root, target, rootReference, _ := mustCatalogSetFixture(t)
	colliding := target.clone()
	colliding.storage = root.storage
	if _, err := NewCatalogSet(
		rootReference, []FieldCatalog{root, colliding},
	); err == nil {
		t.Fatal("resource table reused by a different logical resource")
	}
}

func TestCatalogSetRejectsResourceRelationTableCollision(t *testing.T) {
	_, root, target, rootReference, _ := mustCatalogSetFixture(t)
	colliding := target.clone()
	colliding.storage = ResourceStorage{
		table: root.relations[0].storage.table, validated: true,
	}
	if colliding.Validate() != nil {
		t.Fatal("collision fixture is not an independently valid catalog")
	}
	if _, err := NewCatalogSet(
		rootReference, []FieldCatalog{root, colliding},
	); err == nil {
		t.Fatal("resource and relation shared one physical table")
	}
}

func TestCatalogSetRejectsRelationTableAcrossSourceCatalogs(t *testing.T) {
	_, root, target, rootReference, _ := mustCatalogSetFixture(t)
	relation, err := NewRelationDefinition(
		rootRelationID, "reverse", RelationCardinalityToOne,
		root.relations[0].storage, rootReference, true, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	colliding := target.clone()
	colliding.relations = []RelationDefinition{relation}
	if colliding.Validate() != nil {
		t.Fatal("cross-catalog collision fixture is not independently valid")
	}
	if _, err := NewCatalogSet(
		rootReference, []FieldCatalog{root, colliding},
	); err == nil {
		t.Fatal("relation table reused by a different source catalog")
	}
}
