package recordkernel

import (
	"bytes"
	"errors"
	"github.com/google/uuid"
	"testing"
)

func TestCreateDefaultsThenMaterializesAbsentAndExplicitNull(t *testing.T) {
	catalog, nameID, optionalID, relationID := materializationCatalog(t)
	absent, err := DecodeInput(catalog, ActionCreate,
		fixtureActorAuthority(), []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	absentResult, err := ExecuteMutation(catalog, Record{}, absent)
	if err != nil {
		t.Fatal(err)
	}
	name, _ := absentResult.Record().Value(nameID)
	optional, _ := absentResult.Record().Value(optionalID)
	relation, _ := absentResult.Record().Value(relationID)
	if string(name.CanonicalJSON()) != `"default"` || !optional.IsNull() ||
		string(relation.CanonicalJSON()) != `[]` {
		t.Fatalf("materialized = %s / %s / %s", name.CanonicalJSON(),
			optional.CanonicalJSON(), relation.CanonicalJSON())
	}
	explicit, err := DecodeInput(catalog, ActionCreate,
		fixtureActorAuthority(), []byte(`{"name":null,"links":null}`))
	if err != nil {
		t.Fatal(err)
	}
	explicitResult, err := ExecuteMutation(catalog, Record{}, explicit)
	if err != nil {
		t.Fatal(err)
	}
	name, _ = explicitResult.Record().Value(nameID)
	relation, _ = explicitResult.Record().Value(relationID)
	if !name.IsNull() || string(relation.CanonicalJSON()) != `[]` {
		t.Fatalf("explicit null = %s / %s", name.CanonicalJSON(), relation.CanonicalJSON())
	}
}

func TestMaterializedRecordPreservesCanonicalHashAcrossEmptyUpdate(t *testing.T) {
	catalog, nameID, _, _ := materializationCatalog(t)
	create, _ := DecodeInput(catalog, ActionCreate,
		fixtureActorAuthority(), []byte(`{"name":"kept"}`))
	created, err := ExecuteMutation(catalog, Record{}, create)
	if err != nil {
		t.Fatal(err)
	}
	update, _ := DecodeInput(catalog, ActionUpdate,
		fixtureActorAuthority(), []byte(`{}`))
	updated, err := ExecuteMutation(catalog, created.Record(), update)
	if err != nil || !bytes.Equal(created.CanonicalJSON(), updated.CanonicalJSON()) {
		t.Fatalf("hash continuity failed: %v", err)
	}
	partialValue, _ := created.Record().Value(nameID)
	partial := NewRecord(map[uuid.UUID]Value{nameID: partialValue})
	if _, err := ExecuteMutation(catalog, partial, update); !errors.Is(err, ErrInputInvalid) {
		t.Fatalf("partial persisted record = %v", err)
	}
}

func materializationCatalog(t *testing.T) (Catalog, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	nameID, optionalID, relationID := fixtureV7(601), fixtureV7(602), fixtureV7(603)
	fields := []Field{
		{ID: nameID, PublicName: "name", TypeID: TypeText,
			Role: RoleStored, Nullable: true, WritePolicy: WriteActor},
		{ID: optionalID, PublicName: "optional", TypeID: TypeText,
			Role: RoleStored, Nullable: true, WritePolicy: WriteActor},
		{ID: relationID, PublicName: "links", TypeID: TypeRelation,
			Role: RoleRelation, Nullable: true, WritePolicy: WriteActor,
			Relation: &RelationProfile{Cardinality: "many"}},
	}
	catalog := fixtureCatalog(fields)
	typedDefault, err := buildText("default")
	if err != nil {
		t.Fatal(err)
	}
	catalog.defaults = []defaultValue{{fieldID: nameID,
		expression: expression{kind: "literal", typeID: TypeText,
			literal: typedDefault}}}
	return catalog, nameID, optionalID, relationID
}
