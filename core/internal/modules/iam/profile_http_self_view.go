package iam

import (
	"encoding/json"
	"strings"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func newProfileSelfView(
	record profileapp.SelfRecord,
	schema profileapp.Schema,
	document profileapp.Document,
) (ProfileSelfView, error) {
	if record.Validate(record.ID) != nil || schema.Resource() != profileapp.ResourceUser ||
		schema.State() != profileapp.SchemaActive ||
		document.Resource() != profileapp.ResourceUser {
		return ProfileSelfView{}, ErrProfileHTTPUnavailable
	}
	fields := schema.Fields()
	selected := make([]uuid.UUID, 0, len(fields))
	visible := make(map[uuid.UUID]profileapp.Field, len(fields))
	for _, field := range fields {
		if field.State() != profileapp.FieldActive ||
			field.Classification() == profileapp.ClassificationRestricted {
			continue
		}
		selected = append(selected, field.ID())
		visible[field.ID()] = field
	}
	projected, err := profileapp.ProjectCompatibilityRead(schema, document, selected)
	if err != nil {
		return ProfileSelfView{}, ErrProfileHTTPUnavailable
	}
	var values map[string]json.RawMessage
	if json.Unmarshal(projected, &values) != nil {
		return ProfileSelfView{}, ErrProfileHTTPUnavailable
	}
	result := make(map[string]json.RawMessage, len(values)+12)
	for rawID, value := range values {
		id, parseErr := uuid.Parse(rawID)
		field, found := visible[id]
		if parseErr != nil || !found {
			return ProfileSelfView{}, ErrProfileHTTPUnavailable
		}
		result[field.Key()] = value
	}
	if err := addProfileSelfCore(result, record); err != nil {
		return ProfileSelfView{}, err
	}
	raw, err := json.Marshal(result)
	if err != nil {
		return ProfileSelfView{}, ErrProfileHTTPUnavailable
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, profileapp.MaximumDocumentBytes)
	if err != nil {
		return ProfileSelfView{}, ErrProfileHTTPUnavailable
	}
	return ProfileSelfView{Document: canonical, Revision: record.Revision}, nil
}

func addProfileSelfCore(
	result map[string]json.RawMessage,
	record profileapp.SelfRecord,
) error {
	roleIDs := make([]uuid.UUID, len(record.Roles))
	for index := range record.Roles {
		roleIDs[index] = record.Roles[index].ID
	}
	organizationIDs := make([]uuid.UUID, len(record.Organizations))
	for index := range record.Organizations {
		organizationIDs[index] = record.Organizations[index].ID
	}
	core := map[string]any{
		"id": record.ID, "email": record.Email,
		"status": strings.ToUpper(record.Status), "revision": record.Revision,
		"created_at": record.CreatedAt.UTC(), "updated_at": record.UpdatedAt.UTC(),
		"role_ids": roleIDs, "roles": record.Roles,
		"organization_ids": organizationIDs, "organizations": record.Organizations,
	}
	if len(record.Organizations) > 0 {
		core["organization_id"] = record.Organizations[0].ID
		core["organization"] = record.Organizations[0]
	}
	for key, value := range core {
		encoded, err := json.Marshal(value)
		if err != nil {
			return ErrProfileHTTPUnavailable
		}
		result[key] = encoded
	}
	return nil
}
