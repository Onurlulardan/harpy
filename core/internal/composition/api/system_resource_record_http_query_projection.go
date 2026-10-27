package api

import (
	"encoding/json"
	"strconv"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func projectSystemResourceQuery(
	query systemresources.QueryRecords,
) ([]json.RawMessage, bool, string, *int64, error) {
	items, itemsErr := query.Items()
	hasMore, cursor, exactCount, pageErr := query.Pagination()
	catalog, catalogErr := query.Catalog()
	relations, relationErr := query.Relations()
	if itemsErr != nil || pageErr != nil || catalogErr != nil ||
		relationErr != nil || len(relations) != len(items) {
		return nil, false, "", nil, systemresources.ErrUnavailable
	}
	allowed, err := systemResourceAllowedNames(catalog)
	if err != nil {
		return nil, false, "", nil, err
	}
	result := make([]json.RawMessage, len(items))
	for index, item := range items {
		projected, valid := projectSystemResourceQueryItem(item, allowed)
		if !valid {
			return nil, false, "", nil, systemresources.ErrUnavailable
		}
		result[index], err = projectSystemResourceQueryRelations(
			projected, relations[index],
		)
		if err != nil {
			return nil, false, "", nil, err
		}
	}
	return result, hasMore, cursor, exactCount, nil
}

func projectSystemResourceQueryItem(
	item json.RawMessage,
	allowed map[string]struct{},
) (json.RawMessage, bool) {
	var document map[string]json.RawMessage
	if json.Unmarshal(item, &document) != nil || len(document) == 0 {
		return nil, false
	}
	for name := range document {
		if _, found := allowed[name]; !found || systemResourceSecretName(name) {
			return nil, false
		}
	}
	var id string
	if json.Unmarshal(document["id"], &id) != nil {
		return nil, false
	}
	parsedID, idErr := systemresources.ParseID(id)
	revision, revisionOK := canonicalPositiveRevision(document["revision"])
	if idErr != nil || parsedID.String() != id || !revisionOK {
		return nil, false
	}
	document["revision"] = json.RawMessage(strconv.FormatInt(revision, 10))
	projected, err := json.Marshal(document)
	return projected, err == nil
}

func canonicalPositiveRevision(raw json.RawMessage) (int64, bool) {
	var encoded string
	if json.Unmarshal(raw, &encoded) != nil || encoded == "" {
		return 0, false
	}
	revision, err := strconv.ParseInt(encoded, 10, 64)
	if err != nil || revision < 1 || strconv.FormatInt(revision, 10) != encoded {
		return 0, false
	}
	return revision, true
}
