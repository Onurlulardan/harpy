package api

import "fmt"

func validatePredicateValues(operator Operator, values []Literal) error {
	arity, valid := operator.Arity()
	count := len(values)
	if !valid || arity == PredicateArityZero && count != 0 ||
		arity == PredicateArityOne && count != 1 ||
		arity == PredicateArityList && (count == 0 || count > maximumListValues) ||
		arity == PredicateArityPair && count != 2 {
		return fmt.Errorf("REST Query predicate values do not match operator arity")
	}
	for _, value := range values {
		if !value.valid() {
			return fmt.Errorf("REST Query predicate contains an invalid literal")
		}
	}
	return nil
}

func (filter RequestFilter) valid() bool {
	budget := maximumFilterNodes
	_, _, _, _, valid := filter.measure(0, &budget)
	return valid
}

func (filter RequestFilter) Validate() error {
	if !filter.valid() {
		return fmt.Errorf("REST Query filter is invalid")
	}
	return nil
}

func (filter RequestFilter) measure(
	booleanDepth int, budget *int,
) (uint8, uint16, uint8, uint16, bool) {
	if *budget == 0 {
		return 0, 0, 0, 0, false
	}
	(*budget)--
	if filter.kind == RequestFilterPredicate {
		if filter.predicate == nil {
			return 0, 0, 0, 0, false
		}
		literalCount := uint16(len(filter.predicate.values))
		valid := len(filter.children) == 0 &&
			filter.depth == 0 && filter.nodes == 1 && filter.predicates == 1 &&
			filter.literals == literalCount && filter.predicate.valid()
		return 0, 1, 1, literalCount, valid
	}
	if filter.predicate != nil || booleanDepth >= maximumFilterDepth ||
		!validBooleanChildCount(filter.kind, len(filter.children)) {
		return 0, 0, 0, 0, false
	}
	depth, nodes := uint8(1), uint16(1)
	predicates, literals := uint8(0), uint16(0)
	for _, child := range filter.children {
		childDepth, childNodes, childPredicates, childLiterals, valid :=
			child.measure(booleanDepth+1, budget)
		if !valid {
			return 0, 0, 0, 0, false
		}
		if int(predicates)+int(childPredicates) > maximumFilterPredicates ||
			int(literals)+int(childLiterals) > maximumFilterLiterals {
			return 0, 0, 0, 0, false
		}
		depth = max(depth, childDepth+1)
		nodes += childNodes
		predicates += childPredicates
		literals += childLiterals
	}
	valid := depth == filter.depth && nodes == filter.nodes &&
		predicates == filter.predicates && literals == filter.literals &&
		nodes <= maximumFilterNodes
	return depth, nodes, predicates, literals, valid
}

func validBooleanChildCount(kind RequestFilterKind, count int) bool {
	return (kind == RequestFilterAnd || kind == RequestFilterOr) &&
		count >= 1 && count <= maximumBooleanChildren ||
		kind == RequestFilterNot && count == 1
}

func (predicate RequestPredicate) valid() bool {
	return validRequestName(predicate.field) &&
		validatePredicateValues(predicate.operator, predicate.values) == nil
}

func cloneLiterals(values []Literal) []Literal {
	result := make([]Literal, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}
