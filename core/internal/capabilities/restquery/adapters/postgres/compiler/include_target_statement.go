package compiler

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/google/uuid"
)

type IncludeTargetStatement struct {
	sql           string
	arguments     []any
	columns       []ResultColumn
	stableIndex   int
	stableFieldID uuid.UUID
	class         planner.StatementClass
	catalog       api.CatalogReference
	binding       includeStatementBinding
	stateSeal     [sha256.Size]byte
	validated     bool
}

func newIncludeTargetStatement(
	sql string,
	arguments []any,
	columns []ResultColumn,
	stableIndex int,
	stableFieldID uuid.UUID,
	class planner.StatementClass,
	catalog api.CatalogReference,
	binding includeStatementBinding,
) (IncludeTargetStatement, error) {
	statement := IncludeTargetStatement{
		sql: sql, arguments: cloneArguments(arguments),
		columns: append([]ResultColumn{}, columns...), stableIndex: stableIndex,
		stableFieldID: stableFieldID, class: class, catalog: catalog,
		binding: binding, validated: true,
	}
	seal, err := includeTargetStatementSeal(statement)
	if err != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	statement.stateSeal = seal
	if statement.Validate() != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	return statement, nil
}

func (statement IncludeTargetStatement) Validate() error {
	if !statement.validated || strings.TrimSpace(statement.sql) != statement.sql ||
		statement.sql == "" || statement.class != planner.StatementClassInclude ||
		statement.catalog.Validate() != nil || statement.binding.validate() != nil ||
		statement.binding.role != includeStatementRoleTargetRecord ||
		statement.stableFieldID == uuid.Nil ||
		statement.stableFieldID.Variant() != uuid.RFC4122 ||
		statement.stableIndex < 0 || statement.stableIndex >= len(statement.columns) ||
		len(statement.columns) < 1 || len(statement.columns) > 56 ||
		statement.stateSeal == [sha256.Size]byte{} ||
		validateTargetArguments(statement.arguments) != nil {
		return ErrContract
	}
	for _, column := range statement.columns {
		if column.Validate() != nil {
			return ErrContract
		}
	}
	stableID, err := statement.columns[statement.stableIndex].FieldID()
	seal, sealErr := includeTargetStatementSeal(statement)
	if err != nil || stableID != statement.stableFieldID || sealErr != nil ||
		seal != statement.stateSeal {
		return ErrContract
	}
	return nil
}

func validateTargetArguments(arguments []any) error {
	if len(arguments) < 1 || !validArguments(arguments) {
		return ErrContract
	}
	identifiers, ok := arguments[0].([]uuid.UUID)
	if !ok {
		return ErrContract
	}
	normalized, err := normalizeIncludeUUIDs(identifiers, maximumIncludeCandidateIDs)
	if err != nil || !slices.Equal(identifiers, normalized) {
		return ErrContract
	}
	return nil
}

func (IncludeTargetStatement) String() string   { return compilerRedaction }
func (IncludeTargetStatement) GoString() string { return compilerRedaction }
func (IncludeTargetStatement) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (statement IncludeTargetStatement) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (IncludeTargetStatement) MarshalJSON() ([]byte, error) { return rejectJSON() }
