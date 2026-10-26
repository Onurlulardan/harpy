package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type aggregateTarget struct {
	field     api.FieldDefinition
	fieldType api.FieldType
	alias     string
	hasField  bool
}

func aggregateReferenceTarget(
	input aggregateInput, reference resolved.AggregateReference,
) (aggregateTarget, error) {
	kind, kindErr := reference.Kind()
	ordinal, ordinalErr := reference.Ordinal()
	if kindErr != nil || ordinalErr != nil {
		return aggregateTarget{}, ErrContract
	}
	if kind == api.AggregateReferenceGroup {
		if ordinal < 0 || ordinal >= len(input.groups) {
			return aggregateTarget{}, ErrContract
		}
		fieldType, err := input.groups[ordinal].field.Type()
		alias, aliasErr := aggregateGroupAlias(ordinal)
		if err != nil || aliasErr != nil {
			return aggregateTarget{}, ErrContract
		}
		return aggregateTarget{
			field: input.groups[ordinal].field, fieldType: fieldType,
			alias: alias, hasField: true,
		}, nil
	}
	if kind != api.AggregateReferenceMeasure || ordinal < 0 ||
		ordinal >= len(input.measures) {
		return aggregateTarget{}, ErrContract
	}
	measure := input.measures[ordinal]
	fieldType, err := measure.declaration.ResultType()
	alias, aliasErr := aggregateMeasureAlias(ordinal)
	if err != nil || aliasErr != nil {
		return aggregateTarget{}, ErrContract
	}
	return aggregateTarget{
		field: measure.field, fieldType: fieldType,
		alias: alias, hasField: measure.hasField,
	}, nil
}

func aggregateTargetExpression(
	target aggregateTarget, ordered bool, binder *binder,
) (string, error) {
	kind, err := target.fieldType.Kind()
	if err != nil || target.alias == "" || binder == nil {
		return "", ErrContract
	}
	if kind == api.LogicalTypeEnum && ordered {
		if !target.hasField {
			return "", ErrContract
		}
		values, err := orderedEnumValues(target.field)
		if err != nil {
			return "", err
		}
		codes := make([]string, len(values))
		for index, value := range values {
			codes[index] = value.code
		}
		placeholder, err := binder.bind(codes)
		if err != nil {
			return "", err
		}
		return "pg_catalog.array_position(CAST(" + placeholder +
			" AS pg_catalog.text[]), " + target.alias + ")", nil
	}
	if kind == api.LogicalTypeText {
		return target.alias + " COLLATE pg_catalog.pg_unicode_fast", nil
	}
	return target.alias, nil
}

func bindAggregateTargetValue(
	target aggregateTarget, value resolved.Value,
	ordered bool, binder *binder,
) (string, error) {
	expected, expectedErr := target.fieldType.Kind()
	actual, actualErr := value.Kind()
	if expectedErr != nil || actualErr != nil || expected != actual {
		return "", ErrContract
	}
	if expected == api.LogicalTypeEnum && ordered {
		if !target.hasField {
			return "", ErrContract
		}
		return bindEnumOrdinal(target.field, value, binder)
	}
	return bindResolvedValue(value, binder)
}
