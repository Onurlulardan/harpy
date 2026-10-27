package compiler

import (
	"errors"
	"strings"
	"testing"
)

func TestExplainSQLUsesSealedDeterministicPrefix(t *testing.T) {
	record := fixtureBoundStatement(t, "forms", "display")
	count := fixtureExactCountStatement(t)
	recordSQL, _ := record.SQL()
	countSQL, _ := count.SQL()
	recordExplain, recordErr := record.ExplainSQL()
	countExplain, countErr := count.ExplainSQL()
	if recordErr != nil || countErr != nil ||
		recordExplain != explainJSONPrefix+recordSQL ||
		countExplain != explainJSONPrefix+countSQL ||
		strings.Contains(recordExplain, "ANALYZE") ||
		strings.Contains(countExplain, "ANALYZE") {
		t.Fatal("planner SQL was not compiler-owned deterministic EXPLAIN JSON")
	}
	tampered := record
	tampered.sql += " OR TRUE"
	if _, err := tampered.ExplainSQL(); !errors.Is(err, ErrContract) {
		t.Fatal("tampered record produced planner SQL")
	}
	tamperedCount := count
	tamperedCount.sql += " OR TRUE"
	if _, err := tamperedCount.ExplainSQL(); !errors.Is(err, ErrContract) {
		t.Fatal("tampered count produced planner SQL")
	}
}
