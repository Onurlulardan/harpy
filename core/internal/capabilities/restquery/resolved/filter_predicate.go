package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (predicate Predicate) Validate() error {
	if !predicate.validated || predicate.field.Validate() != nil || predicate.values == nil ||
		predicate.path.Validate() != nil {
		return fmt.Errorf("resolved REST Query predicate is invalid")
	}
	arity, valid := predicate.operator.Arity()
	if !valid || !resolvedArityValid(arity, len(predicate.values)) {
		return fmt.Errorf("resolved REST Query predicate is invalid")
	}
	capabilities, err := predicate.field.Capabilities()
	if err != nil {
		return fmt.Errorf("resolved REST Query predicate is invalid")
	}
	if _, allowed, err := capabilities.Filter(predicate.operator); err != nil || !allowed {
		return fmt.Errorf("resolved REST Query predicate is invalid")
	}
	fieldType, err := predicate.field.Type()
	if err != nil {
		return fmt.Errorf("resolved REST Query predicate is invalid")
	}
	for _, value := range predicate.values {
		if !valueConformsToFieldType(fieldType, predicate.operator, value) {
			return fmt.Errorf("resolved REST Query predicate is invalid")
		}
	}
	return nil
}

func (predicate Predicate) Path() (RequestPath, error) {
	if err := predicate.Validate(); err != nil {
		return RequestPath{}, err
	}
	return predicate.path, nil
}

func (predicate Predicate) Field() (api.FieldDefinition, error) {
	if err := predicate.Validate(); err != nil {
		return api.FieldDefinition{}, err
	}
	return predicate.field, nil
}

func (predicate Predicate) Operator() (api.Operator, error) {
	if err := predicate.Validate(); err != nil {
		return 0, err
	}
	return predicate.operator, nil
}

func (predicate Predicate) Values() ([]Value, error) {
	if err := predicate.Validate(); err != nil {
		return nil, err
	}
	return cloneValues(predicate.values), nil
}

func (predicate Predicate) clone() Predicate {
	result := predicate
	result.values = cloneValues(predicate.values)
	return result
}

func resolvedArityValid(arity api.PredicateArity, count int) bool {
	switch arity {
	case api.PredicateArityZero:
		return count == 0
	case api.PredicateArityOne:
		return count == 1
	case api.PredicateArityList:
		return count > 0
	case api.PredicateArityPair:
		return count == 2
	default:
		return false
	}
}

func predicateValueKind(kind api.LogicalType) api.LogicalType {
	if kind == api.LogicalTypeTextList {
		return api.LogicalTypeText
	}
	if kind == api.LogicalTypeUUIDList {
		return api.LogicalTypeUUID
	}
	return kind
}

func (Predicate) String() string                           { return valueRedaction }
func (Predicate) GoString() string                         { return valueRedaction }
func (predicate Predicate) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Predicate) MarshalJSON() ([]byte, error)             { return rejectDirectJSONMarshal() }
