package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type ExactCountStatement struct {
	sql       string
	arguments []any
	class     planner.StatementClass
	binding   statementBinding
	stateSeal [sha256.Size]byte
	validated bool
}

func newExactCountStatement(
	sql string,
	arguments []any,
	class planner.StatementClass,
	binding statementBinding,
) (ExactCountStatement, error) {
	statement := ExactCountStatement{
		sql: sql, arguments: cloneArguments(arguments),
		class: class, binding: binding, validated: true,
	}
	seal, err := exactCountStatementSeal(statement)
	if err != nil {
		return ExactCountStatement{}, ErrContract
	}
	statement.stateSeal = seal
	if err := statement.Validate(); err != nil {
		return ExactCountStatement{}, err
	}
	return statement, nil
}

func (statement ExactCountStatement) Validate() error {
	if !statement.validated || strings.TrimSpace(statement.sql) != statement.sql ||
		statement.sql == "" || statement.arguments == nil ||
		statement.class != planner.StatementClassExactCount ||
		!validArguments(statement.arguments) || statement.binding.validate() != nil ||
		statement.stateSeal == [sha256.Size]byte{} {
		return ErrContract
	}
	seal, err := exactCountStatementSeal(statement)
	if err != nil || seal != statement.stateSeal {
		return ErrContract
	}
	return nil
}

func (ExactCountStatement) String() string   { return compilerRedaction }
func (ExactCountStatement) GoString() string { return compilerRedaction }
func (ExactCountStatement) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (statement ExactCountStatement) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (ExactCountStatement) MarshalJSON() ([]byte, error) { return rejectJSON() }
