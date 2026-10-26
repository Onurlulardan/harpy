package resolved

func validRecordFilterBudget(record RecordQuery) bool {
	budget := filterMetrics{valid: true}
	if record.where != nil && !addFilterMetrics(&budget, *record.where) {
		return false
	}
	for _, include := range record.includes {
		if !addIncludeFilterMetrics(&budget, include) {
			return false
		}
	}
	return budget.nodes <= 100 && budget.predicates <= 50 && budget.literals <= 1000
}

func addIncludeFilterMetrics(budget *filterMetrics, include Include) bool {
	if include.where != nil && !addFilterMetrics(budget, *include.where) {
		return false
	}
	for _, child := range include.children {
		if !addIncludeFilterMetrics(budget, child) {
			return false
		}
	}
	return true
}

func addFilterMetrics(budget *filterMetrics, filter Filter) bool {
	metrics := measureFilter(filter, 0)
	if !metrics.valid {
		return false
	}
	budget.nodes += metrics.nodes
	budget.predicates += metrics.predicates
	budget.literals += metrics.literals
	return budget.nodes <= 100 && budget.predicates <= 50 && budget.literals <= 1000
}
