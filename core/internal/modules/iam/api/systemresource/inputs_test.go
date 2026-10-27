package systemresource

import (
	"encoding/json"
	"testing"
)

type validSystemResourceQuery struct{}

func (validSystemResourceQuery) Validate() error { return nil }

func TestMutationInputsEnforceResourceSpecificShape(t *testing.T) {
	fieldID := mustSystemResourceID(t, "a1000000-0000-4000-8000-000000000002")
	values, _ := NewFieldValues(map[ID]json.RawMessage{
		fieldID: []byte(`"user@example.com"`),
	})
	created, err := NewCreateInput(Users, values, []byte("password"), "create-1")
	if err != nil {
		t.Fatal(err)
	}
	created.Destroy()
	if created.Validate() == nil {
		t.Fatal("destroyed user secret remained usable")
	}
	if _, err := NewCreateInput(Roles, values, []byte("secret"), "role-1"); err == nil {
		t.Fatal("role accepted a password")
	}
	if _, err := NewStateInput(
		Roles, mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000099"),
		1, StateInactive, "state-1",
	); err == nil {
		t.Fatal("role accepted user state")
	}
}

func TestQueryInputOwnsValidatedRESTQuery(t *testing.T) {
	input, err := NewQueryInput(Users, validSystemResourceQuery{})
	if err != nil || input.Validate() != nil {
		t.Fatalf("query input error=%v", err)
	}
}
