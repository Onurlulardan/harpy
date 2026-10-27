package planner

import (
	"errors"
	"strings"
	"testing"
)

func TestParseExplainCostRejectsShapeDrift(t *testing.T) {
	t.Parallel()
	tests := []string{
		``, `null`, `{}`, `[]`, `[null]`, `[{},{}]`,
		`[{"Plan":null}]`, `[{"Plan":[]}]`, `[{"Plan":{}}]`,
		`[{"plan":{"Total Cost":1}}]`,
		`[{"Plan":{"Plans":[{"Total Cost":1}]}}]`,
		`[{"Plan":{"Total Cost":1},"Planning Time":1}]`,
		`[{"Plan":{"Total Cost":1}}] []`,
	}
	for _, document := range tests {
		assertExplainContractError(t, []byte(document))
	}
}

func TestParseExplainCostRejectsInvalidCost(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`null`, `true`, `"1"`, `-1`, `-0`, `1e9999`, `NaN`, `Infinity`,
	} {
		document := []byte(`[{"Plan":{"Total Cost":` + raw + `}}]`)
		assertExplainContractError(t, document)
	}
}

func TestParseExplainCostRejectsDuplicateKeysAnywhere(t *testing.T) {
	t.Parallel()
	for _, document := range []string{
		`[{"Plan":{"Total Cost":1},"Plan":{"Total Cost":2}}]`,
		`[{"Plan":{"Total Cost":1,"Total Cost":2}}]`,
		`[{"Plan":{"Total Cost":1,"Plans":[{"Node Type":"X","Node Type":"Y"}]}}]`,
	} {
		assertExplainContractError(t, []byte(document))
	}
}

func TestParseExplainCostRejectsExcessiveNesting(t *testing.T) {
	t.Parallel()
	document := `[{"Plan":{"Total Cost":1,"Extra":` +
		strings.Repeat(`[`, 128) + `0` + strings.Repeat(`]`, 128) + `}}]`
	assertExplainContractError(t, []byte(document))
}

func TestExplainErrorsAreOpaque(t *testing.T) {
	t.Parallel()
	secret := "do-not-leak-planner-payload"
	_, err := ParseExplainCost([]byte(`[{"` + secret + `":1}]`))
	if !errors.Is(err, ErrExplainContract) || strings.Contains(err.Error(), secret) {
		t.Fatalf("EXPLAIN error leaked payload: %v", err)
	}
}

func assertExplainContractError(t testing.TB, document []byte) {
	t.Helper()
	if _, err := ParseExplainCost(document); !errors.Is(err, ErrExplainContract) {
		t.Fatalf("document %q error = %v", document, err)
	}
}
