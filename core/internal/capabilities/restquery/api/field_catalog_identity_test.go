package api

import "testing"

func TestFieldCatalogRejectsDuplicateMemberIdentityOrPublicName(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	target := mustCatalogReferenceFixture(t, "crm", targetResourceID, targetVersionID, "2")
	base := catalogFixtureInput(
		t, root, "root_records", rootStableField, rootSearchField, nil,
	)
	duplicateField := base
	duplicateField.Fields = append([]FieldDefinition{}, base.Fields...)
	duplicateField.Fields = append(duplicateField.Fields, base.Fields[0])
	if _, err := NewFieldCatalog(duplicateField); err == nil {
		t.Fatal("duplicate field identity accepted")
	}
	nameCollision := mustUUIDCatalogField(
		t, "60000000-0000-4000-8000-000000000001", "name",
		"fld_other_id", true, true, 36,
	)
	duplicateName := base
	duplicateName.Fields = append([]FieldDefinition{}, base.Fields...)
	duplicateName.Fields = append(duplicateName.Fields, nameCollision)
	if _, err := NewFieldCatalog(duplicateName); err == nil {
		t.Fatal("duplicate field public name accepted")
	}
	relationNameCollision, _ := NewRelationDefinition(
		rootRelationID, "name", RelationCardinalityToOne,
		mustRelationStorage(t, rootRelationID), target, true, 1,
	)
	fieldRelationName := base
	fieldRelationName.Relations = []RelationDefinition{relationNameCollision}
	if _, err := NewFieldCatalog(fieldRelationName); err == nil {
		t.Fatal("field/relation public name collision accepted")
	}
	relationIDCollision, _ := NewRelationDefinition(
		rootStableField, "customer", RelationCardinalityToOne,
		mustRelationStorage(t, rootStableField), target, true, 1,
	)
	fieldRelationID := base
	fieldRelationID.Relations = []RelationDefinition{relationIDCollision}
	if _, err := NewFieldCatalog(fieldRelationID); err == nil {
		t.Fatal("field/relation stable ID collision accepted")
	}
}

func TestFieldCatalogZeroAndMemberTamperFailClosed(t *testing.T) {
	root := mustCatalogReferenceFixture(t, "forms", rootResourceID, rootVersionID, "1")
	valid := mustFieldCatalogFixture(t, catalogFixtureInput(
		t, root, "root_records", rootStableField, rootSearchField, nil,
	))
	unvalidated := valid.clone()
	unvalidated.validated = false
	missingFields := valid.clone()
	missingFields.fields = nil
	invalidMember := valid.clone()
	invalidMember.fields[0].publicName = "_included"
	invalidProfile := valid.clone()
	invalidProfile.defaultSelect = nil
	for _, catalog := range []FieldCatalog{
		{}, unvalidated, missingFields, invalidMember, invalidProfile,
	} {
		if catalog.Validate() == nil {
			t.Fatal("invalid field catalog validated")
		}
		if _, err := catalog.Fields(); err == nil {
			t.Fatal("invalid field catalog accessor succeeded")
		}
	}
}
