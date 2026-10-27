package systemresource

import (
	"strings"
	"unicode/utf8"
)

const maximumRelationSummaryTextBytes = 2048

type RelationSummary struct {
	id          ID
	displayName string
	details     map[string]string
	validated   bool
}

func NewRelationSummary(
	id ID,
	displayName string,
	details map[string]string,
) (RelationSummary, error) {
	summary := RelationSummary{
		id: id, displayName: displayName,
		details: cloneRelationDetails(details), validated: true,
	}
	if summary.Validate() != nil {
		return RelationSummary{}, ErrContract
	}
	return summary, nil
}

func (summary RelationSummary) Validate() error {
	if !summary.validated || !summary.id.Valid() ||
		!validRelationSummaryText(summary.displayName, false) {
		return ErrContract
	}
	if parsed, err := ParseID(summary.displayName); err == nil && parsed.Valid() {
		return ErrContract
	}
	for key, value := range summary.details {
		if !validRelationDetailKey(key) ||
			!validRelationSummaryText(value, true) {
			return ErrContract
		}
	}
	return nil
}

func (summary RelationSummary) Values() (
	ID,
	string,
	map[string]string,
	error,
) {
	if summary.Validate() != nil {
		return ID{}, "", nil, ErrContract
	}
	return summary.id, summary.displayName,
		cloneRelationDetails(summary.details), nil
}

func cloneRelationDetails(values map[string]string) map[string]string {
	if values == nil {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func validRelationDetailKey(value string) bool {
	switch value {
	case "name", "email", "slug", "status":
		return true
	default:
		return false
	}
}

func validRelationSummaryText(value string, allowEmpty bool) bool {
	if !utf8.ValidString(value) || len(value) > maximumRelationSummaryTextBytes ||
		strings.ContainsRune(value, '\x00') ||
		strings.TrimSpace(value) != value {
		return false
	}
	return allowEmpty || value != ""
}
