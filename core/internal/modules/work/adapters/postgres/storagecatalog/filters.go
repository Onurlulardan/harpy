package storagecatalog

import (
	restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
)

func filters(spec queryFieldSpec) ([]restapi.FilterCapabilityInput, error) {
	names := []string{"eq", "ne", "in", "not_in"}
	switch spec.kind {
	case "timestamp", "date", "integer":
		names = []string{"eq", "ne", "gt", "gte", "lt", "lte", "in", "not_in", "between"}
	case "search_text":
		names = []string{"eq", "ne", "contains", "starts_with", "ends_with", "in", "not_in"}
	case "text_list":
		names = []string{"contains_any", "contains_all"}
	}
	if spec.nullable {
		names = append(names, "is_null", "is_not_null")
	}
	result := make([]restapi.FilterCapabilityInput, len(names))
	for index, name := range names {
		operator, found := restapi.ParseOperator(name)
		if !found {
			return nil, workquery.ErrUnavailable
		}
		result[index] = restapi.FilterCapabilityInput{Operator: operator, CountSafe: true}
	}
	return result, nil
}
