package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type RecordStatement struct {
	sql           string
	arguments     []any
	columns       []ResultColumn
	requestedSize int
	class         planner.StatementClass
	countMode     api.CountMode
	cursorContext cursor.Context
	binding       statementBinding
	stateSeal     [sha256.Size]byte
	validated     bool
}

func newRecordStatement(
	sql string,
	arguments []any,
	columns []ResultColumn,
	requestedSize int,
	class planner.StatementClass,
	countMode api.CountMode,
	cursorContext cursor.Context,
	binding statementBinding,
) (RecordStatement, error) {
	statement := RecordStatement{
		sql: sql, arguments: cloneArguments(arguments),
		columns:       append([]ResultColumn{}, columns...),
		requestedSize: requestedSize, class: class, countMode: countMode,
		cursorContext: cursorContext,
		binding:       binding, validated: true,
	}
	seal, err := statementSeal(statement)
	if err != nil {
		return RecordStatement{}, ErrContract
	}
	statement.stateSeal = seal
	if err := statement.Validate(); err != nil {
		return RecordStatement{}, err
	}
	return statement, nil
}

func (statement RecordStatement) Validate() error {
	if !statement.validated || strings.TrimSpace(statement.sql) != statement.sql ||
		statement.sql == "" || statement.arguments == nil || len(statement.columns) < 1 ||
		len(statement.columns) > 56 ||
		statement.requestedSize < 1 || statement.requestedSize > 500 ||
		statement.class != planner.StatementClassRecord ||
		statement.countMode != api.CountModeNone &&
			statement.countMode != api.CountModeExact ||
		!validArguments(statement.arguments) ||
		statement.binding.validate() != nil || statement.cursorContext.Validate() != nil ||
		statement.stateSeal == [sha256.Size]byte{} {
		return ErrContract
	}
	for _, column := range statement.columns {
		if column.Validate() != nil {
			return ErrContract
		}
	}
	if validateCursorContext(
		statement.cursorContext, statement.binding,
		statement.columns, statement.requestedSize,
	) != nil {
		return ErrContract
	}
	seal, err := statementSeal(statement)
	if err != nil || seal != statement.stateSeal {
		return ErrContract
	}
	return nil
}

func (statement RecordStatement) SQL() (string, error) {
	if err := statement.Validate(); err != nil {
		return "", err
	}
	return statement.sql, nil
}

func (statement RecordStatement) Arguments() ([]any, error) {
	if err := statement.Validate(); err != nil {
		return nil, err
	}
	return cloneArguments(statement.arguments), nil
}

func (statement RecordStatement) Columns() ([]ResultColumn, error) {
	if err := statement.Validate(); err != nil {
		return nil, err
	}
	return append([]ResultColumn{}, statement.columns...), nil
}

func (statement RecordStatement) RequestedSize() (int, error) {
	if err := statement.Validate(); err != nil {
		return 0, err
	}
	return statement.requestedSize, nil
}

func (statement RecordStatement) Class() (planner.StatementClass, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.class, nil
}

func (statement RecordStatement) CountMode() (api.CountMode, error) {
	if statement.Validate() != nil {
		return 0, ErrContract
	}
	return statement.countMode, nil
}

func (statement RecordStatement) CursorContext() (cursor.Context, error) {
	if statement.Validate() != nil {
		return cursor.Context{}, ErrContract
	}
	return statement.cursorContext, nil
}

func (RecordStatement) String() string   { return compilerRedaction }
func (RecordStatement) GoString() string { return compilerRedaction }
func (RecordStatement) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (statement RecordStatement) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (RecordStatement) MarshalJSON() ([]byte, error)             { return rejectJSON() }
