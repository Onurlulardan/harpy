package iam

import (
	"encoding/json"
	"testing"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func TestProfileSelfViewRedactsRestrictedDynamicFields(t *testing.T) {
	firstNameID := uuid.MustParse("71000000-0000-4000-8000-000000000001")
	restrictedID := uuid.MustParse("71000000-0000-4000-8000-000000000002")
	firstName := profileHTTPTestField(
		t, firstNameID, "first_name", profileapp.ClassificationInternal,
	)
	restricted := profileHTTPTestField(
		t, restrictedID, "government_id", profileapp.ClassificationRestricted,
	)
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID:       uuid.MustParse("72000000-0000-4000-8000-000000000001"),
		Resource: profileapp.ResourceUser, Version: 1,
		State: profileapp.SchemaActive, Fields: []profileapp.Field{firstName, restricted},
	})
	if err != nil {
		t.Fatal(err)
	}
	document, err := profileapp.DecodeDocument(
		schema, []byte(`{"71000000-0000-4000-8000-000000000001":"Ada",`+
			`"71000000-0000-4000-8000-000000000002":"secret"}`),
		nil, profileHTTPNormalizer{},
	)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC)
	actorID := uuid.MustParse("73000000-0000-4000-8000-000000000001")
	view, err := newProfileSelfView(profileapp.SelfRecord{
		ID: actorID, Email: "ada@example.com", Status: "active", Revision: 2,
		CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
		Roles: []profileapp.SelfRole{}, Organizations: []profileapp.SelfOrganization{},
	}, schema, document)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]json.RawMessage
	if json.Unmarshal(view.Document, &values) != nil ||
		string(values["first_name"]) != `"Ada"` || len(values["government_id"]) != 0 ||
		len(values["password_hash"]) != 0 || view.Revision != 2 {
		t.Fatalf("self profile view=%s revision=%d", view.Document, view.Revision)
	}
}

func profileHTTPTestField(
	t *testing.T,
	id uuid.UUID,
	key string,
	classification profileapp.Classification,
) profileapp.Field {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: id, Resource: profileapp.ResourceUser, Key: key,
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Classification: classification,
		Normalization:  profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
		Validation: profileapp.ValidationInput{MaximumBytes: 200},
	})
	if err != nil {
		t.Fatal(err)
	}
	return field
}
