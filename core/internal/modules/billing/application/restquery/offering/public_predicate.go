package offering

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
)

func enumPredicate(
	field string,
	operator api.Operator,
	canonical string,
) (restauthorization.MandatoryFilter, error) {
	value, err := restauthorization.NewEnumValue(canonical)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	return predicateFilter(
		field, api.LogicalTypeEnum, operator, []restauthorization.Value{value},
	)
}

func booleanPredicate(
	field string,
	operator api.Operator,
	canonical bool,
) (restauthorization.MandatoryFilter, error) {
	value, err := restauthorization.NewBooleanValue(canonical)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	return predicateFilter(
		field, api.LogicalTypeBoolean, operator, []restauthorization.Value{value},
	)
}

func timestampPredicate(
	field string,
	operator api.Operator,
	instant time.Time,
) (restauthorization.MandatoryFilter, error) {
	canonical := instant.UTC().Format(canonicalTimestamp)
	value, err := restauthorization.NewScalarValue(api.LogicalTypeTimestamp, canonical)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	return predicateFilter(
		field, api.LogicalTypeTimestamp, operator, []restauthorization.Value{value},
	)
}

func predicateFilter(
	field string,
	kind api.LogicalType,
	operator api.Operator,
	values []restauthorization.Value,
) (restauthorization.MandatoryFilter, error) {
	predicate, err := restauthorization.NewMandatoryPredicate(
		field, kind, operator, values,
	)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	filter, err := restauthorization.NewMandatoryPredicateFilter(predicate)
	if err != nil {
		return restauthorization.MandatoryFilter{}, ErrContract
	}
	return filter, nil
}
