package api

import "fmt"

type queryFilterBudget struct {
	nodes      int
	predicates int
	literals   int
}

func validateQueryFilterBudget(
	root *RequestFilter, includes []RequestInclude,
) error {
	budget := queryFilterBudget{}
	if err := budget.add(root); err != nil {
		return err
	}
	for index := range includes {
		if err := budget.addInclude(&includes[index]); err != nil {
			return err
		}
	}
	return nil
}

func validateAggregateQueryBudget(root *RequestFilter, having *RequestHaving) error {
	budget := queryFilterBudget{}
	if err := budget.add(root); err != nil {
		return err
	}
	if having == nil {
		return nil
	}
	if !having.valid() {
		return fmt.Errorf("REST Query aggregate HAVING is invalid")
	}
	budget.nodes += int(having.nodes)
	budget.literals += int(having.literals)
	if budget.nodes > maximumFilterNodes || budget.literals > maximumFilterLiterals {
		return fmt.Errorf("REST Query aggregate predicate budget exceeds its limit")
	}
	return nil
}

func (budget *queryFilterBudget) addInclude(include *RequestInclude) error {
	if err := budget.add(include.where); err != nil {
		return err
	}
	for index := range include.children {
		if err := budget.addInclude(&include.children[index]); err != nil {
			return err
		}
	}
	return nil
}

func (budget *queryFilterBudget) add(filter *RequestFilter) error {
	if filter == nil {
		return nil
	}
	if !filter.valid() {
		return fmt.Errorf("REST Query filter is invalid")
	}
	budget.nodes += int(filter.nodes)
	budget.predicates += int(filter.predicates)
	budget.literals += int(filter.literals)
	if budget.nodes > maximumFilterNodes ||
		budget.predicates > maximumFilterPredicates ||
		budget.literals > maximumFilterLiterals {
		return fmt.Errorf("REST Query filter budget exceeds its limit")
	}
	return nil
}
