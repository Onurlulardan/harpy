package billingv1adoptionapp

import "sort"

func finalizePlan(plan *ScopePlan) {
	sort.Slice(plan.Issues, func(left, right int) bool {
		a, b := plan.Issues[left], plan.Issues[right]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.SourceKey < b.SourceKey
	})
	plan.Eligible = len(plan.Issues) == 0
}
