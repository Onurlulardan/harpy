package api

import (
	"encoding/json"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func projectSystemResourceQueryRelations(
	item json.RawMessage,
	relations systemresources.RelationProjection,
) (json.RawMessage, error) {
	var document map[string]json.RawMessage
	if json.Unmarshal(item, &document) != nil || len(document) == 0 {
		return nil, systemresources.ErrUnavailable
	}
	if err := projectSystemResourceRelations(document, relations); err != nil {
		return nil, err
	}
	projected, err := json.Marshal(document)
	if err != nil {
		return nil, systemresources.ErrUnavailable
	}
	return projected, nil
}

func projectSystemResourceRelations(
	document map[string]json.RawMessage,
	projection systemresources.RelationProjection,
) error {
	_, values, err := projection.Values()
	if err != nil {
		return systemresources.ErrUnavailable
	}
	for _, value := range values {
		field, multiple, summaries, valueErr := value.Values()
		displayKey, detailsKey := field+"_display", field+"_details"
		if valueErr != nil || document[displayKey] != nil ||
			document[detailsKey] != nil {
			return systemresources.ErrUnavailable
		}
		displays, details, detailErr := systemResourceRelationDocuments(summaries)
		if detailErr != nil || !multiple && len(displays) > 1 {
			return systemresources.ErrUnavailable
		}
		if multiple {
			document[displayKey], err = json.Marshal(displays)
			if err == nil {
				document[detailsKey], err = json.Marshal(details)
			}
		} else if len(displays) == 0 {
			document[displayKey], document[detailsKey] =
				json.RawMessage("null"), json.RawMessage("null")
		} else {
			document[displayKey], err = json.Marshal(displays[0])
			if err == nil {
				document[detailsKey], err = json.Marshal(details[0])
			}
		}
		if err != nil {
			return systemresources.ErrUnavailable
		}
	}
	return nil
}

func systemResourceRelationDocuments(
	summaries []systemresources.RelationSummary,
) ([]string, []map[string]string, error) {
	displays := make([]string, len(summaries))
	details := make([]map[string]string, len(summaries))
	for index, summary := range summaries {
		id, display, fields, err := summary.Values()
		if err != nil || fields["id"] != "" || fields["display_name"] != "" {
			return nil, nil, systemresources.ErrUnavailable
		}
		fields["id"], fields["display_name"] = id.String(), display
		displays[index], details[index] = display, fields
	}
	return displays, details, nil
}
