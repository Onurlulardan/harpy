package aggregatepage

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type metadata struct {
	groups   []column
	measures []column
	limit    int
}

type column struct {
	fieldType api.FieldType
	key       string
}

func loadMetadata(statement compiler.AggregateStatement) (metadata, error) {
	class, classErr := statement.Class()
	groups, groupErr := statement.Groups()
	measures, measureErr := statement.Measures()
	limit, limitErr := statement.Limit()
	if classErr != nil || groupErr != nil || measureErr != nil || limitErr != nil ||
		class != planner.StatementClassAggregate {
		return metadata{}, ErrContract
	}
	result := metadata{
		groups:   make([]column, len(groups)),
		measures: make([]column, len(measures)), limit: limit,
	}
	for index, group := range groups {
		fieldType, typeErr := group.FieldType()
		key, keyErr := group.OutputKey()
		if typeErr != nil || keyErr != nil {
			return metadata{}, ErrContract
		}
		result.groups[index] = column{fieldType: fieldType, key: key}
	}
	for index, measure := range measures {
		fieldType, typeErr := measure.FieldType()
		key, keyErr := measure.OutputKey()
		if typeErr != nil || keyErr != nil {
			return metadata{}, ErrContract
		}
		result.measures[index] = column{fieldType: fieldType, key: key}
	}
	if validateMetadata(result) != nil {
		return metadata{}, ErrContract
	}
	return result, nil
}

func validateMetadata(value metadata) error {
	if value.groups == nil || value.measures == nil ||
		len(value.groups) > 3 || len(value.measures) > 8 ||
		len(value.groups)+len(value.measures) < 1 ||
		value.limit < 1 || value.limit > 500 ||
		!validColumns(value.groups, false) || !validColumns(value.measures, true) {
		return ErrContract
	}
	return nil
}

func validColumns(columns []column, measure bool) bool {
	seen := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		kind, err := column.fieldType.Kind()
		if err != nil || column.key == "" ||
			measure && !measureKind(kind) {
			return false
		}
		if _, duplicate := seen[column.key]; duplicate {
			return false
		}
		seen[column.key] = struct{}{}
	}
	return true
}

func measureKind(kind api.LogicalType) bool {
	return kind == api.LogicalTypeText || kind == api.LogicalTypeInteger ||
		kind == api.LogicalTypeDecimal || kind == api.LogicalTypeDate ||
		kind == api.LogicalTypeTimestamp || kind == api.LogicalTypeEnum
}
