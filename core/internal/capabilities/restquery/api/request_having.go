package api

import "fmt"

const (
	maximumHavingDepth      = 3
	maximumHavingPredicates = 10
)

type RequestHavingKind uint8

const (
	RequestHavingKindPredicate RequestHavingKind = iota + 1
	RequestHavingKindAnd
	RequestHavingKindOr
	RequestHavingKindNot
)

type RequestHaving struct {
	kind       RequestHavingKind
	predicate  *RequestHavingPredicate
	children   []RequestHaving
	depth      uint8
	nodes      uint16
	predicates uint8
	literals   uint16
}

type RequestHavingPredicate struct {
	reference AggregateReference
	name      string
	operator  Operator
	values    []Literal
}

func NewRequestHavingPredicate(
	reference AggregateReference, name string, operator Operator, values []Literal,
) (RequestHaving, error) {
	predicate := RequestHavingPredicate{
		reference: reference, name: name, operator: operator,
		values: cloneLiterals(values),
	}
	if !predicate.valid() {
		return RequestHaving{}, fmt.Errorf("REST Query HAVING predicate is invalid")
	}
	return RequestHaving{
		kind: RequestHavingKindPredicate, predicate: &predicate,
		nodes: 1, predicates: 1, literals: uint16(len(values)),
	}, nil
}

func NewRequestHavingAnd(children []RequestHaving) (RequestHaving, error) {
	return newRequestHavingBoolean(RequestHavingKindAnd, children)
}

func NewRequestHavingOr(children []RequestHaving) (RequestHaving, error) {
	return newRequestHavingBoolean(RequestHavingKindOr, children)
}

func NewRequestHavingNot(child RequestHaving) (RequestHaving, error) {
	return newRequestHavingBoolean(RequestHavingKindNot, []RequestHaving{child})
}

func (having RequestHaving) Kind() RequestHavingKind { return having.kind }
func (having RequestHaving) Depth() int              { return int(having.depth) }
func (having RequestHaving) NodeCount() int          { return int(having.nodes) }
func (having RequestHaving) PredicateCount() int     { return int(having.predicates) }
func (having RequestHaving) Predicate() (RequestHavingPredicate, bool) {
	if having.kind != RequestHavingKindPredicate || having.predicate == nil {
		return RequestHavingPredicate{}, false
	}
	return having.predicate.clone(), true
}
func (having RequestHaving) Children() []RequestHaving {
	return cloneRequestHavingSlice(having.children)
}

func (having RequestHaving) String() string                 { return queryValueRedaction }
func (having RequestHaving) GoString() string               { return queryValueRedaction }
func (having RequestHaving) Format(state fmt.State, _ rune) { writeQueryRedaction(state) }
func (RequestHaving) MarshalJSON() ([]byte, error)          { return rejectDirectJSONMarshal() }
func (RequestHavingKind) MarshalJSON() ([]byte, error)      { return rejectDirectJSONMarshal() }

func (predicate RequestHavingPredicate) Reference() AggregateReference {
	return predicate.reference
}
func (predicate RequestHavingPredicate) Name() string       { return predicate.name }
func (predicate RequestHavingPredicate) Operator() Operator { return predicate.operator }
func (predicate RequestHavingPredicate) Values() []Literal  { return cloneLiterals(predicate.values) }
func (predicate RequestHavingPredicate) String() string     { return queryValueRedaction }
func (predicate RequestHavingPredicate) GoString() string   { return queryValueRedaction }
func (predicate RequestHavingPredicate) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestHavingPredicate) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}

func (having RequestHaving) clone() RequestHaving {
	result := having
	result.children = cloneRequestHavingSlice(having.children)
	if having.predicate != nil {
		predicate := having.predicate.clone()
		result.predicate = &predicate
	}
	return result
}

func (predicate RequestHavingPredicate) clone() RequestHavingPredicate {
	result := predicate
	result.values = cloneLiterals(predicate.values)
	return result
}

func cloneRequestHavingSlice(values []RequestHaving) []RequestHaving {
	result := make([]RequestHaving, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}
