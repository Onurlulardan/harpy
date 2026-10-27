package api

import (
	"testing"

	"github.com/google/uuid"
)

func TestFieldCatalogRoundTripAndPublicNameResolution(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	target := mustCatalogReferenceFixture(t, "crm", targetResourceID, targetVersionID, "2")
	relation, err := NewRelationDefinition(
		rootRelationID, "customer", RelationCardinalityToOne,
		mustRelationStorage(t, rootRelationID), target, true, 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	input := catalogFixtureInput(
		t, root, "root_records", rootStableField, rootSearchField,
		[]RelationDefinition{relation},
	)
	catalog := mustFieldCatalogFixture(t, input)
	actualReference, referenceErr := catalog.Reference()
	storage, storageErr := catalog.Storage()
	capabilities, capabilitiesErr := catalog.Capabilities()
	fields, fieldsErr := catalog.Fields()
	relations, relationsErr := catalog.Relations()
	defaults, defaultsErr := catalog.DefaultSelect()
	stableID, stableErr := catalog.StableID()
	search, searchErr := catalog.SearchFields()
	if referenceErr != nil || storageErr != nil || capabilitiesErr != nil ||
		fieldsErr != nil || relationsErr != nil || defaultsErr != nil ||
		stableErr != nil || searchErr != nil {
		t.Fatal("valid field catalog accessor failed")
	}
	if actualReference != root || storage.Validate() != nil || capabilities.Validate() != nil ||
		len(fields) != 2 || len(relations) != 1 || len(defaults) != 2 ||
		stableID != uuid.MustParse(rootStableField) ||
		len(search) != 1 || search[0] != uuid.MustParse(rootSearchField) {
		t.Fatal("field catalog did not round trip")
	}
	field, found, err := catalog.FieldByPublicName("name")
	if err != nil || !found || field.publicName != "name" {
		t.Fatal("field public name did not resolve")
	}
	actualRelation, found, err := catalog.RelationByPublicName("customer")
	if err != nil || !found || actualRelation.relationID != relation.relationID {
		t.Fatal("relation public name did not resolve")
	}
	if _, found, err := catalog.FieldByPublicName("unknown"); err != nil || found {
		t.Fatal("unknown valid field name did not return a clean miss")
	}
	if _, found, err := catalog.RelationByPublicName("unknown"); err != nil || found {
		t.Fatal("unknown valid relation name did not return a clean miss")
	}
	for _, invalid := range []string{"", "e\u0301", "bad\x00name", "_included"} {
		if _, _, err := catalog.FieldByPublicName(invalid); err == nil {
			t.Fatalf("invalid field lookup accepted: %q", invalid)
		}
		if _, _, err := catalog.RelationByPublicName(invalid); err == nil {
			t.Fatalf("invalid relation lookup accepted: %q", invalid)
		}
	}
}

func TestFieldCatalogDeepImmutability(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	target := mustCatalogReferenceFixture(t, "crm", targetResourceID, targetVersionID, "2")
	relation, _ := NewRelationDefinition(
		rootRelationID, "customer", RelationCardinalityToOne,
		mustRelationStorage(t, rootRelationID), target, true, 1,
	)
	input := catalogFixtureInput(
		t, root, "root_records", rootStableField, rootSearchField,
		[]RelationDefinition{relation},
	)
	catalog := mustFieldCatalogFixture(t, input)
	input.Fields[0].publicName = "mutated"
	input.Relations[0].publicName = "mutated_relation"
	input.DefaultSelect[0] = uuid.Nil
	input.SearchFields[0] = uuid.Nil
	if catalog.Validate() != nil {
		t.Fatal("field catalog retained constructor input storage")
	}
	fields, _ := catalog.Fields()
	relations, _ := catalog.Relations()
	defaults, _ := catalog.DefaultSelect()
	search, _ := catalog.SearchFields()
	fields[0].publicName = "mutated"
	fields[0].capabilities.filters[0] = FilterCapability{}
	relations[0].publicName = "mutated_relation"
	defaults[0] = uuid.Nil
	search[0] = uuid.Nil
	if catalog.Validate() != nil {
		t.Fatal("field catalog accessor exposed internal storage")
	}
	if _, found, _ := catalog.Field(uuid.MustParse(rootStableField)); !found {
		t.Fatal("field ID lookup was mutated")
	}
	if _, found, _ := catalog.Relation(uuid.MustParse(rootRelationID)); !found {
		t.Fatal("relation ID lookup was mutated")
	}
}
