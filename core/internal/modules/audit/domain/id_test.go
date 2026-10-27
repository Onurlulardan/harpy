package domain

import "testing"

func TestIDAcceptsAdoptedUUIDv4AndClassifiesNewUUIDv7(t *testing.T) {
	legacy, err := ParseID("85f17fb4-d9f0-4ed8-99ee-7c079ba863fa")
	if err != nil || !legacy.Valid() || legacy.UUIDv7() {
		t.Fatalf("legacy ID=%s v7=%v err=%v", legacy.String(), legacy.UUIDv7(), err)
	}
	created, err := ParseID("019c0000-0000-7000-8000-000000000701")
	if err != nil || !created.UUIDv7() {
		t.Fatalf("created ID=%s v7=%v err=%v", created.String(), created.UUIDv7(), err)
	}
	for _, invalid := range []string{
		"00000000-0000-0000-0000-000000000000",
		"85F17FB4-D9F0-4ED8-99EE-7C079BA863FA",
		"not-an-id",
	} {
		if _, err := ParseID(invalid); err == nil {
			t.Fatalf("invalid ID %q was accepted", invalid)
		}
	}
}
