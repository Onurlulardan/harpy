package iam

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func identityRelationSummary(
	resource publicresource.Resource,
	item json.RawMessage,
	allowed map[string]struct{},
) (uuid.UUID, publicresource.RelationSummary, error) {
	var document map[string]json.RawMessage
	if json.Unmarshal(item, &document) != nil || document == nil {
		return uuid.Nil, publicresource.RelationSummary{},
			publicresource.ErrUnavailable
	}
	for field := range document {
		if _, found := allowed[field]; !found {
			return uuid.Nil, publicresource.RelationSummary{},
				publicresource.ErrUnavailable
		}
	}
	id, err := identityRelationDocumentID(document["id"])
	if err != nil {
		return uuid.Nil, publicresource.RelationSummary{}, err
	}
	texts, err := identityRelationDocumentTexts(document)
	if err != nil {
		return uuid.Nil, publicresource.RelationSummary{}, err
	}
	display := texts["name"]
	if resource == publicresource.Users {
		display = texts["display_name"]
		if display == "" {
			display = texts["email"]
		}
	}
	if display == "" {
		return uuid.Nil, publicresource.RelationSummary{},
			publicresource.ErrUnavailable
	}
	details := make(map[string]string, 4)
	for _, field := range []string{"name", "email", "slug", "status"} {
		if value := texts[field]; value != "" {
			details[field] = value
		}
	}
	summary, err := publicresource.NewRelationSummary(
		publicresource.ID(id), display, details,
	)
	if err != nil {
		return uuid.Nil, publicresource.RelationSummary{},
			publicresource.ErrUnavailable
	}
	return id, summary, nil
}

func identityRelationDocumentID(raw json.RawMessage) (uuid.UUID, error) {
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return uuid.Nil, publicresource.ErrUnavailable
	}
	id, err := uuid.Parse(value)
	if err != nil || id.String() != value {
		return uuid.Nil, publicresource.ErrUnavailable
	}
	return id, nil
}

func identityRelationDocumentTexts(
	document map[string]json.RawMessage,
) (map[string]string, error) {
	result := make(map[string]string, len(document))
	for field, raw := range document {
		if field == "id" || string(raw) == "null" {
			continue
		}
		var value string
		if json.Unmarshal(raw, &value) != nil {
			return nil, publicresource.ErrUnavailable
		}
		result[field] = value
	}
	return result, nil
}
