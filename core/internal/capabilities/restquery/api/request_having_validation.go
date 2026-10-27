package api

import "fmt"

func newRequestHavingBoolean(
	kind RequestHavingKind, children []RequestHaving,
) (RequestHaving, error) {
	if !validHavingChildCount(kind, len(children)) {
		return RequestHaving{}, fmt.Errorf("REST Query HAVING boolean child count is invalid")
	}
	result := RequestHaving{
		kind: kind, children: make([]RequestHaving, len(children)), depth: 1, nodes: 1,
	}
	for index, child := range children {
		if !child.valid() || int(child.depth)+1 > maximumHavingDepth ||
			int(result.nodes)+int(child.nodes) > maximumFilterNodes ||
			int(result.predicates)+int(child.predicates) > maximumHavingPredicates ||
			int(result.literals)+int(child.literals) > maximumFilterLiterals {
			return RequestHaving{}, fmt.Errorf("REST Query HAVING structure exceeds its limits")
		}
		result.children[index] = child.clone()
		result.depth = max(result.depth, child.depth+1)
		result.nodes += child.nodes
		result.predicates += child.predicates
		result.literals += child.literals
	}
	return result, nil
}

func (having RequestHaving) valid() bool {
	nodeBudget, predicateBudget := maximumFilterNodes, maximumHavingPredicates
	_, _, _, _, valid := having.measure(0, &nodeBudget, &predicateBudget)
	return valid
}

func (having RequestHaving) measure(
	booleanDepth int, nodeBudget *int, predicateBudget *int,
) (uint8, uint16, uint8, uint16, bool) {
	if *nodeBudget == 0 {
		return 0, 0, 0, 0, false
	}
	(*nodeBudget)--
	if having.kind == RequestHavingKindPredicate {
		if *predicateBudget == 0 || having.predicate == nil {
			return 0, 0, 0, 0, false
		}
		(*predicateBudget)--
		literalCount := uint16(len(having.predicate.values))
		valid := len(having.children) == 0 && having.depth == 0 &&
			having.nodes == 1 && having.predicates == 1 &&
			having.literals == literalCount && having.predicate.valid()
		return 0, 1, 1, literalCount, valid
	}
	if having.predicate != nil || booleanDepth >= maximumHavingDepth ||
		!validHavingChildCount(having.kind, len(having.children)) {
		return 0, 0, 0, 0, false
	}
	depth, nodes := uint8(1), uint16(1)
	predicates, literals := uint8(0), uint16(0)
	for _, child := range having.children {
		childDepth, childNodes, childPredicates, childLiterals, valid :=
			child.measure(booleanDepth+1, nodeBudget, predicateBudget)
		if !valid || int(predicates)+int(childPredicates) > maximumHavingPredicates ||
			int(literals)+int(childLiterals) > maximumFilterLiterals {
			return 0, 0, 0, 0, false
		}
		depth = max(depth, childDepth+1)
		nodes += childNodes
		predicates += childPredicates
		literals += childLiterals
	}
	valid := depth == having.depth && nodes == having.nodes &&
		predicates == having.predicates && literals == having.literals
	return depth, nodes, predicates, literals, valid
}

func validHavingChildCount(kind RequestHavingKind, count int) bool {
	return (kind == RequestHavingKindAnd || kind == RequestHavingKindOr) &&
		count >= 1 && count <= maximumBooleanChildren ||
		kind == RequestHavingKindNot && count == 1
}

func (predicate RequestHavingPredicate) valid() bool {
	if !predicate.reference.valid() ||
		!validAggregateReferenceName(predicate.reference, predicate.name) {
		return false
	}
	return validatePredicateValues(predicate.operator, predicate.values) == nil
}

func validAggregateReferenceName(reference AggregateReference, name string) bool {
	if reference == AggregateReferenceGroup {
		return validRequestName(name)
	}
	return reference == AggregateReferenceMeasure && validMeasureName(name)
}

func (having RequestHaving) referencesDeclared(
	groups map[string]struct{}, measures map[string]struct{},
) bool {
	if having.kind == RequestHavingKindPredicate {
		_, groupFound := groups[having.predicate.name]
		_, measureFound := measures[having.predicate.name]
		return having.predicate.reference == AggregateReferenceGroup && groupFound ||
			having.predicate.reference == AggregateReferenceMeasure && measureFound
	}
	for _, child := range having.children {
		if !child.referencesDeclared(groups, measures) {
			return false
		}
	}
	return true
}
