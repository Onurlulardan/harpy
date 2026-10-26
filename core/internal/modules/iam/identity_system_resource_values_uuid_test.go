package iam

import (
	"encoding/json"
	"fmt"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func TestIdentitySystemResourceUUIDListCanonicalizesExplicitSet(t *testing.T) {
	high := "a2000000-0000-4000-8000-000000000002"
	low := "a2000000-0000-4000-8000-000000000001"
	values := identitySystemResourceValues{values: map[publicresource.ID]json.RawMessage{
		mustIdentitySystemResourceID(t, publicresource.UserOrganizationIDsFieldID): json.RawMessage(`["` + high + `","` + low + `"]`),
	}}
	ids, found, err := values.uuidList(publicresource.UserOrganizationIDsFieldID, false)
	if err != nil || !found || len(ids) != 2 || ids[0].String() != low || ids[1].String() != high {
		t.Fatalf("canonical UUID set = %v/%t/%v", ids, found, err)
	}
}

func TestIdentitySystemResourceUUIDListEnforcesBudgetAndUniqueness(t *testing.T) {
	fieldID := mustIdentitySystemResourceID(t, publicresource.UserRoleIDsFieldID)
	values := make([]string, 101)
	for index := range values {
		values[index] = fmt.Sprintf("a3000000-0000-4000-8000-%012x", index+1)
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	input := identitySystemResourceValues{values: map[publicresource.ID]json.RawMessage{
		fieldID: encoded,
	}}
	if _, _, err = input.uuidList(publicresource.UserRoleIDsFieldID, false); err == nil {
		t.Fatal("101 relation IDs were accepted")
	}
	duplicate, _ := json.Marshal([]string{values[0], values[0]})
	input.values[fieldID] = duplicate
	if _, _, err = input.uuidList(publicresource.UserRoleIDsFieldID, false); err == nil {
		t.Fatal("duplicate relation IDs were accepted")
	}
	empty := identitySystemResourceValues{values: map[publicresource.ID]json.RawMessage{
		fieldID: json.RawMessage(`[]`),
	}}
	ids, found, err := empty.uuidList(publicresource.UserRoleIDsFieldID, false)
	if err != nil || !found || len(ids) != 0 {
		t.Fatalf("explicit empty UUID set = %#v/found=%t/error=%v", ids, found, err)
	}
}

func mustIdentitySystemResourceID(t *testing.T, value string) publicresource.ID {
	t.Helper()
	id, err := publicresource.ParseID(value)
	if err != nil || uuid.UUID(id) == uuid.Nil {
		t.Fatalf("parse ID %q: %v", value, err)
	}
	return id
}
