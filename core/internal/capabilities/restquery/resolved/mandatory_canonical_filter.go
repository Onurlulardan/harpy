package resolved

import (
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type mandatoryPredicateDocument struct {
	Kind        string            `json:"kind"`
	FieldID     string            `json:"field_id"`
	LogicalType string            `json:"logical_type"`
	Operator    string            `json:"operator"`
	Values      []json.RawMessage `json:"values"`
}

type mandatoryBooleanDocument struct {
	Kind     string            `json:"kind"`
	Children []json.RawMessage `json:"children"`
}

func canonicalMandatoryFilter(filter MandatoryFilter) (json.RawMessage, error) {
	if filter.Validate() != nil {
		return nil, fmt.Errorf("resolved REST Query mandatory filter is invalid")
	}
	if filter.predicate != nil {
		return canonicalMandatoryPredicate(*filter.predicate)
	}
	children := make([]json.RawMessage, 0, len(filter.children))
	for _, child := range filter.children {
		canonical, err := canonicalMandatoryFilter(child)
		if err != nil {
			return nil, err
		}
		children = append(children, canonical)
	}
	if filter.kind == api.RequestFilterAnd || filter.kind == api.RequestFilterOr {
		sortCanonicalDocuments(children)
	}
	document := mandatoryBooleanDocument{
		Kind: mandatoryFilterKindName(filter.kind), Children: children,
	}
	return attestation.MarshalCanonical(document)
}

func canonicalMandatoryPredicate(
	predicate MandatoryPredicate,
) (json.RawMessage, error) {
	if predicate.Validate() != nil {
		return nil, fmt.Errorf("resolved REST Query mandatory predicate is invalid")
	}
	identifier, idErr := predicate.field.ID()
	fieldType, typeErr := predicate.field.Type()
	kind, kindErr := fieldType.Kind()
	if idErr != nil || typeErr != nil || kindErr != nil {
		return nil, fmt.Errorf("resolved REST Query mandatory predicate is invalid")
	}
	values := make([]json.RawMessage, 0, len(predicate.values))
	for _, value := range predicate.values {
		document, err := canonicalMandatoryValue(value)
		if err != nil {
			return nil, err
		}
		values = append(values, document)
	}
	if setValuedOperator(predicate.operator) {
		sortCanonicalDocuments(values)
	}
	document := mandatoryPredicateDocument{
		Kind: "predicate", FieldID: identifier.String(),
		LogicalType: kind.String(), Operator: predicate.operator.String(),
		Values: values,
	}
	return attestation.MarshalCanonical(document)
}

func mandatoryFilterKindName(kind api.RequestFilterKind) string {
	switch kind {
	case api.RequestFilterAnd:
		return "and"
	case api.RequestFilterOr:
		return "or"
	case api.RequestFilterNot:
		return "not"
	default:
		return ""
	}
}
