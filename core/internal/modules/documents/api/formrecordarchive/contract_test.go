package formrecordarchive

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateRequestRequiresStableUniqueLinkedItems(t *testing.T) {
	definition := contractID("018f0000-0000-7000-8000-000000000101")
	record := contractID("018f0000-0000-7000-8000-000000000102")
	item := contractID("018f0000-0000-7000-8000-000000000103")
	request := CreateRequest{DefinitionID: definition, RecordID: record,
		ItemIDs: []ID{item}, IdempotencyKey: "record-archive-1"}
	if !request.Valid() {
		t.Fatal("valid form-record archive request was rejected")
	}
	request.ItemIDs = []ID{item, item}
	if request.Valid() {
		t.Fatal("duplicate linked item was accepted")
	}
	request.ItemIDs = []ID{item}
	request.RecordID = ID{}
	if request.Valid() {
		t.Fatal("missing record identity was accepted")
	}
}

func TestAccessPinsActorTenantAndAuthorizationRevisions(t *testing.T) {
	actor := contractID("018f0000-0000-7000-8000-000000000111")
	tenant := contractID("018f0000-0000-7000-8000-000000000112")
	access, err := NewAccess(actor, tenant, 3, 5, 7)
	if err != nil || !access.Valid() || access.ActorID() != actor ||
		access.TenantID() != tenant || access.TokenVersion() != 3 ||
		access.AuthorizationVersion() != 5 || access.AuthorizationClock() != 7 {
		t.Fatalf("access evidence was not pinned: %#v err=%v", access, err)
	}
	if _, err = NewAccess(actor, tenant, 3, 0, 7); err == nil {
		t.Fatal("zero authorization revision was accepted")
	}
}

func contractID(value string) ID { return ID(uuid.MustParse(value)) }
