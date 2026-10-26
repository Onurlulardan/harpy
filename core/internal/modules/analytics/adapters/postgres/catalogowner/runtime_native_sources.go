package catalogowner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/httpdecode"
	nativequery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/nativequery"
)

func nativeJoinSources(
	definition endpointDefinition, fields []nativequery.Field,
) ([]nativequery.Source, error) {
	filters, err := nativeJoinFilters(definition)
	if err != nil {
		return nil, err
	}
	result := make([]nativequery.Source, len(definition.DataSources))
	for index, source := range definition.DataSources {
		selected := make([]string, 0, len(fields)+2)
		seen := map[string]bool{}
		add := func(field string) {
			if field != "" && !seen[field] {
				seen[field], selected = true, append(selected, field)
			}
		}
		for _, field := range fields {
			if field.Source == source.Alias {
				add(field.Field)
			}
		}
		for _, join := range definition.Joins {
			if join.SourceForm == source.Alias {
				field, ok := unqualified(join.SourceField, source.Alias)
				if !ok {
					return nil, fmt.Errorf("Analytics native source join field is invalid")
				}
				add(field)
			}
			if join.TargetForm == source.Alias {
				field, ok := unqualified(join.TargetField, source.Alias)
				if !ok {
					return nil, fmt.Errorf("Analytics native target join field is invalid")
				}
				add(field)
			}
		}
		where, err := endpointWhere(filters[source.Alias], source.Alias)
		if err != nil || len(selected) == 0 {
			return nil, fmt.Errorf("Analytics native source query is invalid")
		}
		query := map[string]any{"select": selected,
			"page":  map[string]any{"size": 50, "after": nil},
			"count": "none", "include": []any{}}
		if where != nil {
			query["where"] = where
		}
		document, err := json.Marshal(query)
		if err != nil {
			return nil, fmt.Errorf("encode Analytics native source query: %w", err)
		}
		if _, err = httpdecode.POST().Decode(document); err != nil {
			return nil, fmt.Errorf("validate Analytics native source query: %w", err)
		}
		result[index] = nativequery.Source{Alias: source.Alias,
			FormSlug: source.FormSlug, Query: document}
	}
	return result, nil
}

func nativeJoinFilters(
	definition endpointDefinition,
) (map[string][]endpointFilter, error) {
	sources := make(map[string]bool, len(definition.DataSources))
	for _, source := range definition.DataSources {
		sources[source.Alias] = true
	}
	result := make(map[string][]endpointFilter, len(sources))
	for _, filter := range definition.Filters {
		delimiter := strings.IndexByte(filter.Field, '.')
		if delimiter < 1 || delimiter == len(filter.Field)-1 {
			return nil, fmt.Errorf("Analytics native source filter is ambiguous")
		}
		alias := filter.Field[:delimiter]
		if !sources[alias] {
			return nil, fmt.Errorf("Analytics native source filter is invalid")
		}
		result[alias] = append(result[alias], filter)
	}
	return result, nil
}
