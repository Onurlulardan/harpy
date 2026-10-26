package systemresource

import (
	"encoding/json"
	"testing"
)

func TestRelationProjectionOwnsTypedHumanReadableSummaries(t *testing.T) {
	roleID := mustSystemResourceID(
		t, "018f0f4e-7b61-7c25-8000-000000000071",
	)
	summary, err := NewRelationSummary(
		roleID, "Tenant administrators",
		map[string]string{"name": "Tenant administrators", "status": "active"},
	)
	if err != nil {
		t.Fatal(err)
	}
	value, err := NewToManyRelation("role_ids", []RelationSummary{summary})
	if err != nil {
		t.Fatal(err)
	}
	projection, err := NewRelationProjection(Users, []RelationValue{value})
	if err != nil {
		t.Fatal(err)
	}
	resource, values, err := projection.Values()
	if err != nil || resource != Users || len(values) != 1 {
		t.Fatalf("projection resource=%s values=%d error=%v", resource, len(values), err)
	}
	field, multiple, summaries, err := values[0].Values()
	if err != nil || field != "role_ids" || !multiple || len(summaries) != 1 {
		t.Fatalf("relation field=%s multiple=%v summaries=%d error=%v",
			field, multiple, len(summaries), err)
	}
	id, display, details, err := summaries[0].Values()
	if err != nil || id != roleID || display != "Tenant administrators" ||
		details["status"] != "active" {
		t.Fatalf("summary id=%s display=%q details=%v error=%v",
			id.String(), display, details, err)
	}
	details["status"] = "retired"
	_, _, owned, _ := summaries[0].Values()
	if owned["status"] != "active" {
		t.Fatal("relation summary details were not owned")
	}
}

func TestRelationProjectionRejectsTechnicalDisplayAndWrongResourceField(
	t *testing.T,
) {
	id := mustSystemResourceID(
		t, "018f0f4e-7b61-7c25-8000-000000000072",
	)
	if _, err := NewRelationSummary(
		id, id.String(), map[string]string{"name": "safe"},
	); err == nil {
		t.Fatal("UUID display value was accepted")
	}
	summary, err := NewRelationSummary(
		id, "Harpy", map[string]string{"slug": "harpy"},
	)
	if err != nil {
		t.Fatal(err)
	}
	parent, err := NewToOneRelation("parent_id", &summary)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = NewRelationProjection(Users, []RelationValue{parent}); err == nil {
		t.Fatal("organization relation was accepted for users")
	}
}

func TestReadAndQueryResultsBindRelationProjectionToResource(t *testing.T) {
	catalog := testUserFieldCatalog(t)
	userID := mustSystemResourceID(
		t, "018f0f4e-7b61-7c25-8000-000000000073",
	)
	values, err := NewFieldValues(map[ID]json.RawMessage{
		mustSystemResourceID(t, UserIDFieldID): json.RawMessage(`"` + userID.String() + `"`),
	})
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewRecord(Users, userID, 1, values)
	if err != nil {
		t.Fatal(err)
	}
	relations, err := EmptyRelationProjection(Users)
	if err != nil {
		t.Fatal(err)
	}
	read, err := NewReadResultWithRelations(record, catalog, relations)
	if err != nil {
		t.Fatal(err)
	}
	if projected, err := read.Relations(); err != nil || projected.Validate() != nil {
		t.Fatalf("read relations invalid: %v", err)
	}
	item := json.RawMessage(`{"id":"` + userID.String() + `"}`)
	query, err := NewQueryResultWithRelations(
		[]json.RawMessage{item}, false, "", nil, catalog,
		[]RelationProjection{relations},
	)
	if err != nil {
		t.Fatal(err)
	}
	projected, err := query.Relations()
	if err != nil || len(projected) != 1 || projected[0].Validate() != nil {
		t.Fatalf("query relations invalid: count=%d error=%v", len(projected), err)
	}
}
