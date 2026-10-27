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

type RelationEdgeStatement struct {
	sql       string
	arguments []any
	class     planner.StatementClass
	catalog   api.CatalogReference
	binding   includeStatementBinding
	stateSeal [sha256.Size]byte
	validated bool
}

func newRelationEdgeStatement(
	sql string,
	arguments []any,
	class planner.StatementClass,
	catalog api.CatalogReference,
	binding includeStatementBinding,
) (RelationEdgeStatement, error) {
	statement := RelationEdgeStatement{
		sql: sql, arguments: cloneArguments(arguments), class: class,
		catalog: catalog, binding: binding, validated: true,
	}
	seal, err := relationEdgeStatementSeal(statement)
	if err != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	statement.stateSeal = seal
	if statement.Validate() != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	return statement, nil
}

func (statement RelationEdgeStatement) Validate() error {
	if !statement.validated || strings.TrimSpace(statement.sql) != statement.sql ||
		statement.sql == "" || statement.class != planner.StatementClassInclude ||
		statement.catalog.Validate() != nil || statement.binding.validate() != nil ||
		statement.binding.role != includeStatementRoleRelationEdge ||
		statement.stateSeal == [sha256.Size]byte{} ||
		validateRelationArguments(statement.arguments) != nil {
		return ErrContract
	}
	seal, err := relationEdgeStatementSeal(statement)
	if err != nil || seal != statement.stateSeal {
		return ErrContract
	}
	return nil
}

func validateRelationArguments(arguments []any) error {
	if len(arguments) != 1 {
		return ErrContract
	}
	identifiers, ok := arguments[0].([]uuid.UUID)
	if !ok {
		return ErrContract
	}
	normalized, err := normalizeIncludeUUIDs(identifiers, maximumIncludeSourceIDs)
	if err != nil || !slices.Equal(identifiers, normalized) {
		return ErrContract
	}
	return nil
}

func (RelationEdgeStatement) String() string   { return compilerRedaction }
func (RelationEdgeStatement) GoString() string { return compilerRedaction }
func (RelationEdgeStatement) LogValue() slog.Value {
	return slog.StringValue(compilerRedaction)
}
func (statement RelationEdgeStatement) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (RelationEdgeStatement) MarshalJSON() ([]byte, error) { return rejectJSON() }
