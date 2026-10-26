package resolved

import "testing"

func TestValidateIncludeTreeCountsOnlyVisibleProjections(t *testing.T) {
	root, include := resolvedBudgetInclude(t, 50)
	projections, edges, valid := validateIncludeTree(
		[]Include{include}, root, map[resourceNodeKey]struct{}{}, 1,
	)
	if !valid || projections != 50 || edges != 1 {
		t.Fatalf("include tree = projections:%d edges:%d valid:%v", projections, edges, valid)
	}
	root, include = resolvedBudgetInclude(t, 51)
	projections, _, valid = validateIncludeTree(
		[]Include{include}, root, map[resourceNodeKey]struct{}{}, 1,
	)
	if !valid || projections != 51 {
		t.Fatalf("51-visible tree = projections:%d valid:%v", projections, valid)
	}
}
