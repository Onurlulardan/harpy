package api

import "testing"

func TestRelationDefinitionIncludePolicy(t *testing.T) {
	target := mustRelationTarget(t)
	valid := []struct {
		cardinality RelationCardinality
		enabled     bool
		maximum     int
	}{
		{RelationCardinalityToOne, false, 0},
		{RelationCardinalityToMany, false, 0},
		{RelationCardinalityToOne, true, 1},
		{RelationCardinalityToMany, true, 1},
		{RelationCardinalityToMany, true, 100},
	}
	for _, policy := range valid {
		if _, err := NewRelationDefinition(
			validRelationID, "customer", policy.cardinality,
			mustRelationStorage(t, validRelationID), target, policy.enabled, policy.maximum,
		); err != nil {
			t.Fatalf("valid include policy rejected: %+v", policy)
		}
	}
	invalid := []struct {
		cardinality RelationCardinality
		enabled     bool
		maximum     int
	}{
		{RelationCardinalityToOne, false, 1},
		{RelationCardinalityToMany, false, 1},
		{RelationCardinalityToOne, true, 0},
		{RelationCardinalityToOne, true, 2},
		{RelationCardinalityToMany, true, 0},
		{RelationCardinalityToMany, true, 101},
	}
	for _, policy := range invalid {
		if _, err := NewRelationDefinition(
			validRelationID, "customer", policy.cardinality,
			mustRelationStorage(t, validRelationID), target, policy.enabled, policy.maximum,
		); err == nil {
			t.Fatalf("invalid include policy accepted: %+v", policy)
		}
	}
}

func TestRelationDefinitionZeroAccessorsFailClosed(t *testing.T) {
	var definition RelationDefinition
	if definition.Validate() == nil {
		t.Fatal("zero relation definition validated")
	}
	checks := []func() error{
		func() error { _, err := definition.RelationID(); return err },
		func() error { _, err := definition.PublicName(); return err },
		func() error { _, err := definition.Cardinality(); return err },
		func() error { _, err := definition.Storage(); return err },
		func() error { _, err := definition.Target(); return err },
		func() error { _, err := definition.IncludeEnabled(); return err },
		func() error { _, err := definition.MaximumIncludedRecords(); return err },
	}
	for _, check := range checks {
		if check() == nil {
			t.Fatal("zero relation definition accessor succeeded")
		}
	}
}

func TestRelationDefinitionPackageTamperFailsClosed(t *testing.T) {
	target := mustRelationTarget(t)
	valid, _ := NewRelationDefinition(
		validRelationID, "customer", RelationCardinalityToMany,
		mustRelationStorage(t, validRelationID), target, true, 10,
	)
	unparsed := valid
	unparsed.parsed = false
	missingID := valid
	missingID.relationID = [16]byte{}
	invalidName := valid
	invalidName.publicName = "_included"
	invalidCardinality := valid
	invalidCardinality.cardinality = 0
	invalidStorage := valid
	invalidStorage.storage.stateSeal[0] ^= 1
	invalidTarget := valid
	invalidTarget.target = CatalogReference{}
	invalidPolicy := valid
	invalidPolicy.maximumIncludedRecords = 101
	for _, definition := range []RelationDefinition{
		unparsed, missingID, invalidName, invalidCardinality, invalidStorage,
		invalidTarget, invalidPolicy,
	} {
		if definition.Validate() == nil {
			t.Fatal("tampered relation definition validated")
		}
	}
}
