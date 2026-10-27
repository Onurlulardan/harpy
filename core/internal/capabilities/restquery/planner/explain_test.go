package planner

import (
	"errors"
	"strings"
	"testing"
)

func TestParseExplainCostReadsOnlyRootTotalCost(t *testing.T) {
	t.Parallel()
	document := []byte(`[{"Plan":{"Node Type":"Nested Loop","Startup Cost":0.1,` +
		`"Total Cost":49999.75,"Plan Rows":3,"Plans":[` +
		`{"Node Type":"Seq Scan","Total Cost":999999.99}]}}]`)
	cost, err := ParseExplainCost(document)
	if err != nil || cost.Value() != 49_999.75 {
		t.Fatalf("root cost = %s, %v", cost, err)
	}
}

func TestParseExplainCostAcceptsFiniteJSONNumbers(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		raw  string
		want float64
	}{
		{"0", 0},
		{"50000", 50_000},
		{"5e4", 50_000},
		{"0.125", 0.125},
	} {
		document := []byte(`[{"Plan":{"Total Cost":` + test.raw + `}}]`)
		cost, err := ParseExplainCost(document)
		if err != nil || cost.Value() != test.want {
			t.Fatalf("cost %s = %s, %v", test.raw, cost, err)
		}
	}
}

func TestParseExplainCostEnforcesRawByteBoundary(t *testing.T) {
	t.Parallel()
	prefix := `[{"Plan":{"Total Cost":1}}]`
	atLimit := []byte(prefix + strings.Repeat(" ", MaximumExplainBytes-len(prefix)))
	if _, err := ParseExplainCost(atLimit); err != nil {
		t.Fatalf("maximum document rejected: %v", err)
	}
	overLimit := append(append([]byte(nil), atLimit...), ' ')
	if _, err := ParseExplainCost(overLimit); !errors.Is(err, ErrExplainContract) {
		t.Fatalf("oversized document error = %v", err)
	}
}

func TestCostIsImmutableValue(t *testing.T) {
	t.Parallel()
	original, _ := ParseExplainCost([]byte(`[{"Plan":{"Total Cost":12.5}}]`))
	copyOfCost := original
	copyOfCost.value = 99
	if original.Value() != 12.5 {
		t.Fatalf("original cost = %s", original)
	}
}
