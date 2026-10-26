package planner

import "testing"

func FuzzParseExplainCost(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`[{"Plan":{"Total Cost":1}}]`),
		[]byte(`[{"Plan":{"Total Cost":1,"Total Cost":2}}]`),
		[]byte(`[{"Plan":{"Total Cost":1e9999}}]`),
		nil,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, document []byte) {
		cost, err := ParseExplainCost(document)
		if err == nil && !cost.valid() {
			t.Fatalf("accepted invalid cost: %s", cost)
		}
	})
}
