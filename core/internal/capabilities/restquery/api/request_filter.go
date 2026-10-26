package api

import (
	"fmt"
)

const (
	maximumBooleanChildren  = 50
	maximumFilterDepth      = 5
	maximumFilterNodes      = 100
	maximumFilterPredicates = 50
	maximumFilterLiterals   = 1000
	maximumListValues       = 100
)

func (RequestFilterKind) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

type RequestFilterKind uint8

const (
	RequestFilterPredicate RequestFilterKind = iota + 1
	RequestFilterAnd
	RequestFilterOr
	RequestFilterNot
)

type RequestFilter struct {
	kind       RequestFilterKind
	predicate  *RequestPredicate
	children   []RequestFilter
	depth      uint8
	nodes      uint16
	predicates uint8
	literals   uint16
}

type RequestPredicate struct {
	field    string
	operator Operator
	values   []Literal
}

func NewRequestPredicate(
	field string, operator Operator, values []Literal,
) (RequestFilter, error) {
	if !validRequestName(field) {
		return RequestFilter{}, fmt.Errorf("REST Query field name is invalid")
	}
	if err := validatePredicateValues(operator, values); err != nil {
		return RequestFilter{}, err
	}
	predicate := &RequestPredicate{
		field: field, operator: operator, values: cloneLiterals(values),
	}
	return RequestFilter{
		kind: RequestFilterPredicate, predicate: predicate, nodes: 1,
		predicates: 1, literals: uint16(len(values)),
	}, nil
}

func NewRequestAnd(children []RequestFilter) (RequestFilter, error) {
	return newRequestBoolean(RequestFilterAnd, children)
}

func NewRequestOr(children []RequestFilter) (RequestFilter, error) {
	return newRequestBoolean(RequestFilterOr, children)
}

func NewRequestNot(child RequestFilter) (RequestFilter, error) {
	return newRequestBoolean(RequestFilterNot, []RequestFilter{child})
}

func (filter RequestFilter) Kind() RequestFilterKind { return filter.kind }
func (filter RequestFilter) Depth() int              { return int(filter.depth) }
func (filter RequestFilter) NodeCount() int          { return int(filter.nodes) }

func (filter RequestFilter) Predicate() (RequestPredicate, bool) {
	if filter.kind != RequestFilterPredicate || filter.predicate == nil {
		return RequestPredicate{}, false
	}
	return filter.predicate.clone(), true
}

func (filter RequestFilter) Children() []RequestFilter {
	children := make([]RequestFilter, len(filter.children))
	for index := range filter.children {
		children[index] = filter.children[index].clone()
	}
	return children
}

func (filter RequestFilter) String() string   { return queryValueRedaction }
func (filter RequestFilter) GoString() string { return queryValueRedaction }
func (filter RequestFilter) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestFilter) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (predicate RequestPredicate) Field() string                  { return predicate.field }
func (predicate RequestPredicate) Operator() Operator             { return predicate.operator }
func (predicate RequestPredicate) Values() []Literal              { return cloneLiterals(predicate.values) }
func (predicate RequestPredicate) String() string                 { return queryValueRedaction }
func (predicate RequestPredicate) GoString() string               { return queryValueRedaction }
func (predicate RequestPredicate) Format(state fmt.State, _ rune) { writeQueryRedaction(state) }
func (RequestPredicate) MarshalJSON() ([]byte, error)             { return rejectDirectJSONMarshal() }

func newRequestBoolean(kind RequestFilterKind, children []RequestFilter) (RequestFilter, error) {
	if !validBooleanChildCount(kind, len(children)) {
		return RequestFilter{}, fmt.Errorf("REST Query boolean child count is invalid")
	}
	result := RequestFilter{kind: kind, children: make([]RequestFilter, len(children)), depth: 1, nodes: 1}
	for index, child := range children {
		if !child.valid() || int(child.depth)+1 > maximumFilterDepth ||
			int(result.nodes)+int(child.nodes) > maximumFilterNodes ||
			int(result.predicates)+int(child.predicates) > maximumFilterPredicates ||
			int(result.literals)+int(child.literals) > maximumFilterLiterals {
			return RequestFilter{}, fmt.Errorf("REST Query filter structure exceeds its limits")
		}
		result.children[index] = child.clone()
		result.nodes += child.nodes
		result.predicates += child.predicates
		result.literals += child.literals
		result.depth = max(result.depth, child.depth+1)
	}
	return result, nil
}

func (filter RequestFilter) clone() RequestFilter {
	result := filter
	if filter.predicate != nil {
		predicate := filter.predicate.clone()
		result.predicate = &predicate
	}
	result.children = filter.Children()
	return result
}

func (predicate RequestPredicate) clone() RequestPredicate {
	return RequestPredicate{
		field: predicate.field, operator: predicate.operator,
		values: cloneLiterals(predicate.values),
	}
}
