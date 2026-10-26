package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func fixturePredecessorStatement(
	t *testing.T, input recordInput, columns []recordColumn, index int64,
) PredecessorStatement {
	t.Helper()
	binder := newBinder()
	sql, err := compileV1PredecessorSQL(input, columns, binder, index)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := v1PredecessorBinding(input.binding, index)
	if err != nil {
		t.Fatal(err)
	}
	statement, err := newPredecessorStatement(
		sql, binder.arguments, resultColumns(columns), index,
		planner.StatementClassV1PredecessorSeek, binding,
	)
	if err != nil {
		t.Fatal(err)
	}
	return statement
}
