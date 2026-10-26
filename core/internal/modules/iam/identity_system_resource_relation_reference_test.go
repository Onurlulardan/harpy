package iam

import (
	"encoding/json"
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestIdentityRelationReferencesPreserveDeclaredOrderAndNull(t *testing.T) {
	first := "018f0f4e-7b61-7c25-8000-000000000091"
	second := "018f0f4e-7b61-7c25-8000-000000000092"
	references, err := identityRelationReferences(
		publicresource.Users,
		json.RawMessage(`{"role_ids":["`+second+`","`+first+
			`"],"organization_ids":[]}`),
	)
	if err != nil || len(references) != 2 ||
		references[0].definition.field != "role_ids" ||
		len(references[0].ids) != 2 ||
		references[0].ids[0].String() != second ||
		references[0].ids[1].String() != first ||
		references[1].definition.field != "organization_ids" ||
		len(references[1].ids) != 0 {
		t.Fatalf("references=%+v error=%v", references, err)
	}
	parent, err := identityRelationReferences(
		publicresource.Organizations,
		json.RawMessage(`{"parent_id":null}`),
	)
	if err != nil || len(parent) != 1 || len(parent[0].ids) != 0 {
		t.Fatalf("null parent references=%+v error=%v", parent, err)
	}
	organization, err := identityRelationReferences(
		publicresource.Organizations,
		json.RawMessage(`{"user_ids":["`+second+
			`"],"owner_ids":["`+first+`"]}`),
	)
	if err != nil || len(organization) != 2 ||
		organization[0].definition.field != "user_ids" ||
		organization[1].definition.field != "owner_ids" ||
		organization[1].definition.target != publicresource.Users ||
		organization[1].ids[0].String() != first {
		t.Fatalf("organization references=%+v error=%v", organization, err)
	}
}

func TestIdentityRelationReferencesRejectInvalidAndDuplicateIdentity(t *testing.T) {
	id := "018f0f4e-7b61-7c25-8000-000000000093"
	for _, item := range []json.RawMessage{
		json.RawMessage(`{"role_ids":["not-a-uuid"]}`),
		json.RawMessage(`{"role_ids":["` + id + `","` + id + `"]}`),
		json.RawMessage(`{"organization_ids":null}`),
	} {
		if _, err := identityRelationReferences(
			publicresource.Users, item,
		); !errors.Is(err, publicresource.ErrUnavailable) &&
			!errors.Is(err, publicresource.ErrContract) {
			t.Fatalf("item=%s error=%v", item, err)
		}
	}
}
