package api

import (
	"encoding/json"
	"errors"
	"testing"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func TestSystemResourceRelationProjectionAddsDisplayAndTypedDetails(t *testing.T) {
	roleID, err := systemresources.ParseID(
		"018f0f4e-7b61-7c25-8000-000000000081",
	)
	if err != nil {
		t.Fatal(err)
	}
	summary, err := systemresources.NewRelationSummary(
		roleID, "Tenant administrators",
		map[string]string{"name": "Tenant administrators", "status": "active"},
	)
	if err != nil {
		t.Fatal(err)
	}
	relation, err := systemresources.NewToManyRelation(
		"role_ids", []systemresources.RelationSummary{summary},
	)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := systemresources.NewRelationProjection(
		systemresources.Users, []systemresources.RelationValue{relation},
	)
	if err != nil {
		t.Fatal(err)
	}
	item := json.RawMessage(`{"role_ids":["` + roleID.String() + `"]}`)
	projected, err := projectSystemResourceQueryRelations(item, projection)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if json.Unmarshal(projected, &document) != nil {
		t.Fatal("query relation projection is not valid JSON")
	}
	var displays []string
	var details []map[string]string
	if json.Unmarshal(document["role_ids_display"], &displays) != nil ||
		json.Unmarshal(document["role_ids_details"], &details) != nil ||
		len(displays) != 1 || displays[0] != "Tenant administrators" ||
		len(details) != 1 || details[0]["id"] != roleID.String() ||
		details[0]["display_name"] != "Tenant administrators" {
		t.Fatalf("unexpected relation projection: display=%v details=%v",
			displays, details)
	}
}

func TestSystemResourceRelationProjectionFailsClosedOnWireCollision(t *testing.T) {
	projection, err := systemresources.NewRelationProjection(
		systemresources.Users, []systemresources.RelationValue{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := projectSystemResourceRelations(
		map[string]json.RawMessage{"role_ids": json.RawMessage(`[]`)},
		projection,
	); err != nil {
		t.Fatal(err)
	}
	summaryID, _ := systemresources.ParseID(
		"018f0f4e-7b61-7c25-8000-000000000082",
	)
	summary, _ := systemresources.NewRelationSummary(
		summaryID, "Role", map[string]string{"name": "Role"},
	)
	relation, _ := systemresources.NewToManyRelation(
		"role_ids", []systemresources.RelationSummary{summary},
	)
	projection, _ = systemresources.NewRelationProjection(
		systemresources.Users, []systemresources.RelationValue{relation},
	)
	document := map[string]json.RawMessage{
		"role_ids":         json.RawMessage(`[]`),
		"role_ids_display": json.RawMessage(`"attacker-controlled"`),
	}
	if err := projectSystemResourceRelations(document, projection); !errors.Is(
		err, systemresources.ErrUnavailable,
	) {
		t.Fatalf("collision error=%v", err)
	}
}

func TestSystemResourceRelationProjectionAddsOrganizationGetCompanions(
	t *testing.T,
) {
	ownerID, _ := systemresources.ParseID(
		"018f0f4e-7b61-7c25-8000-000000000083",
	)
	memberID, _ := systemresources.ParseID(
		"018f0f4e-7b61-7c25-8000-000000000084",
	)
	roleID, _ := systemresources.ParseID(
		"018f0f4e-7b61-7c25-8000-000000000085",
	)
	owner, _ := systemresources.NewRelationSummary(
		ownerID, "Ada Lovelace", map[string]string{"email": "ada@example.com"},
	)
	member, _ := systemresources.NewRelationSummary(
		memberID, "Grace Hopper", map[string]string{"email": "grace@example.com"},
	)
	role, _ := systemresources.NewRelationSummary(
		roleID, "Administrators", map[string]string{"name": "Administrators"},
	)
	owners, _ := systemresources.NewToManyRelation(
		"owner_ids", []systemresources.RelationSummary{owner},
	)
	members, _ := systemresources.NewToManyRelation(
		"user_ids", []systemresources.RelationSummary{owner, member},
	)
	roles, _ := systemresources.NewToManyRelation(
		"role_ids", []systemresources.RelationSummary{role},
	)
	projection, _ := systemresources.NewRelationProjection(
		systemresources.Organizations,
		[]systemresources.RelationValue{members, roles, owners},
	)
	document := map[string]json.RawMessage{
		"id": json.RawMessage(`"` + ownerID.String() + `"`),
	}
	if err := projectSystemResourceRelations(document, projection); err != nil {
		t.Fatal(err)
	}
	var displays []string
	if json.Unmarshal(document["owner_ids_display"], &displays) != nil ||
		len(displays) != 1 || displays[0] != "Ada Lovelace" ||
		document["owner_ids"] != nil || document["user_ids"] != nil ||
		document["role_ids"] != nil ||
		document["user_ids_display"] == nil || document["role_ids_display"] == nil {
		t.Fatalf("unexpected organization relation document: %v", document)
	}
}
