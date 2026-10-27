package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

// PredecessorStatement is the sealed V1 compatibility seek statement.
type PredecessorStatement struct {
	sql              string
	arguments        []any
	columns          []ResultColumn
	predecessorIndex int64
	class            planner.StatementClass
	binding          statementBinding
	stateSeal        [sha256.Size]byte
	validated        bool
}

func newPredecessorStatement(
	sql string,
	arguments []any,
	columns []ResultColumn,
	index int64,
	class planner.StatementClass,
	binding statementBinding,
) (PredecessorStatement, error) {
	statement := PredecessorStatement{
		sql: sql, arguments: cloneArguments(arguments),
		columns: append([]ResultColumn{}, columns...), predecessorIndex: index,
		class: class, binding: binding, validated: true,
	}
	seal, err := predecessorStatementSeal(statement)
	if err != nil {
		return PredecessorStatement{}, ErrContract
	}
	statement.stateSeal = seal
	if statement.Validate() != nil {
		return PredecessorStatement{}, ErrContract
	}
	return statement, nil
}

func (statement PredecessorStatement) Validate() error {
	if !statement.validated || strings.TrimSpace(statement.sql) != statement.sql ||
		statement.sql == "" || statement.arguments == nil ||
		statement.predecessorIndex < 0 ||
		statement.class != planner.StatementClassV1PredecessorSeek ||
		statement.binding.validate() != nil || !validArguments(statement.arguments) ||
		statement.stateSeal == [sha256.Size]byte{} ||
		validatePredecessorColumns(statement.columns) != nil {
		return ErrContract
	}
	seal, err := predecessorStatementSeal(statement)
	if err != nil || seal != statement.stateSeal {
		return ErrContract
	}
	return nil
}

func validatePredecessorColumns(columns []ResultColumn) error {
	if len(columns) < 1 || len(columns) > 6 {
		return ErrContract
	}
	for index, column := range columns {
		ordinal, sorted, ordinalErr := column.SortOrdinal()
		_, hasOutput, outputErr := column.OutputKey()
		if ordinalErr != nil || outputErr != nil || !sorted || hasOutput ||
			ordinal != index {
			return ErrContract
		}
	}
	return nil
}

func (PredecessorStatement) String() string   { return compilerRedaction }
func (PredecessorStatement) GoString() string { return compilerRedaction }
func (PredecessorStatement) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (statement PredecessorStatement) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (PredecessorStatement) MarshalJSON() ([]byte, error) { return rejectJSON() }
