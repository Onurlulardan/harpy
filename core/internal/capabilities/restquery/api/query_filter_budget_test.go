package api

import "testing"

func TestRecordQueryRequestEnforcesGlobalPredicateBudget(t *testing.T) {
	root := mustPredicateFilterCount(t, 30, 1)
	includeWhere := mustPredicateFilterCount(t, 20, 1)
	input := recordInputWithIncludeWhere(t, root, includeWhere)
	if _, err := NewRecordQueryRequest(input); err != nil {
		t.Fatalf("50 predicates rejected: %v", err)
	}
	includeWhere = mustPredicateFilterCount(t, 21, 1)
	input = recordInputWithIncludeWhere(t, root, includeWhere)
	if _, err := NewRecordQueryRequest(input); err == nil {
		t.Fatal("51 predicates across root/include accepted")
	}
}

func TestRecordQueryRequestEnforcesGlobalASTNodeBudget(t *testing.T) {
	root := mustWrappedPredicateFilter(t, 25)
	includeWhere := mustWrappedPredicateFilter(t, 25)
	input := recordInputWithIncludeWhere(t, root, includeWhere)
	if _, err := NewRecordQueryRequest(input); err == nil {
		t.Fatal("102 AST nodes across root/include accepted")
	}
}

func TestRecordQueryRequestEnforcesGlobalLiteralBudget(t *testing.T) {
	root := mustPredicateFilterCount(t, 5, 100)
	includeWhere := mustPredicateFilterCount(t, 5, 100)
	input := recordInputWithIncludeWhere(t, root, includeWhere)
	if _, err := NewRecordQueryRequest(input); err != nil {
		t.Fatalf("1000 literals rejected: %v", err)
	}
	includeWhere = mustPredicateFilterCount(t, 6, 100)
	input = recordInputWithIncludeWhere(t, root, includeWhere)
	if _, err := NewRecordQueryRequest(input); err == nil {
		t.Fatal("1100 literals across root/include accepted")
	}
}

func recordInputWithIncludeWhere(
	t *testing.T, root RequestFilter, includeWhere RequestFilter,
) RecordQueryRequestInput {
	t.Helper()
	include := mustRequestInclude(t, RequestIncludeInput{
		Relation: "child", Where: &includeWhere,
	})
	return RecordQueryRequestInput{
		Where: &root, Include: []RequestInclude{include},
		Page: mustRequestPage(t, 50, nil), Count: CountModeNone,
	}
}

func mustPredicateFilterCount(t *testing.T, count int, literalCount int) RequestFilter {
	t.Helper()
	literal := mustLiteral(t, `"value"`)
	values := make([]Literal, literalCount)
	for index := range values {
		values[index] = literal
	}
	predicate, err := NewRequestPredicate("field", OperatorIn, values)
	if err != nil {
		t.Fatal(err)
	}
	filter, err := NewRequestAnd(repeatFilters(predicate, count))
	if err != nil {
		t.Fatal(err)
	}
	return filter
}

func mustWrappedPredicateFilter(t *testing.T, count int) RequestFilter {
	t.Helper()
	filter, err := NewRequestAnd(wrappedPredicates(t, mustPredicate(t), count))
	if err != nil {
		t.Fatal(err)
	}
	return filter
}
