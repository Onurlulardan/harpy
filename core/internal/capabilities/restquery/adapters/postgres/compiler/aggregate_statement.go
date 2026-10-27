package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type AggregateStatement struct {
	sql       string
	arguments []any
	groups    []AggregateGroupColumn
	measures  []AggregateMeasureColumn
	limit     int
	class     planner.StatementClass
	binding   statementBinding
	stateSeal [sha256.Size]byte
	validated bool
}

func newAggregateStatement(
	sql string, arguments []any,
	groups []AggregateGroupColumn,
	measures []AggregateMeasureColumn,
	limit int, binding statementBinding,
) (AggregateStatement, error) {
	statement := AggregateStatement{
		sql: sql, arguments: cloneArguments(arguments),
		groups:   append([]AggregateGroupColumn{}, groups...),
		measures: append([]AggregateMeasureColumn{}, measures...),
		limit:    limit, class: planner.StatementClassAggregate,
		binding: binding, validated: true,
	}
	seal, err := aggregateStatementSeal(statement)
	if err != nil {
		return AggregateStatement{}, ErrContract
	}
	statement.stateSeal = seal
	if statement.Validate() != nil {
		return AggregateStatement{}, ErrContract
	}
	return statement, nil
}

func (statement AggregateStatement) Validate() error {
	if !statement.validated || strings.TrimSpace(statement.sql) != statement.sql ||
		statement.sql == "" || statement.arguments == nil ||
		len(statement.groups) > 3 || len(statement.measures) > 8 ||
		len(statement.groups)+len(statement.measures) < 1 ||
		statement.limit < 1 || statement.limit > 500 ||
		statement.class != planner.StatementClassAggregate ||
		!validArguments(statement.arguments) || statement.binding.validate() != nil ||
		statement.stateSeal == [sha256.Size]byte{} {
		return ErrContract
	}
	for _, group := range statement.groups {
		if group.Validate() != nil {
			return ErrContract
		}
	}
	for _, measure := range statement.measures {
		if measure.Validate() != nil {
			return ErrContract
		}
	}
	seal, err := aggregateStatementSeal(statement)
	if err != nil || seal != statement.stateSeal {
		return ErrContract
	}
	return nil
}

func (AggregateStatement) String() string   { return compilerRedaction }
func (AggregateStatement) GoString() string { return compilerRedaction }
func (AggregateStatement) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (statement AggregateStatement) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (AggregateStatement) MarshalJSON() ([]byte, error) { return rejectJSON() }
