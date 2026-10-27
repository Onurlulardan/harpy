package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (compiler Compiler) CompileRelationEdge(
	plan resolved.AuthorizedPlan,
	graph IncludeGraph,
	node IncludeNode,
	sourceIDs []uuid.UUID,
) (RelationEdgeStatement, error) {
	if compiler.Validate() != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	input, err := includeInputForPlan(plan, graph, node)
	if err != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	identifiers, err := normalizeIncludeUUIDs(sourceIDs, maximumIncludeSourceIDs)
	if err != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	storage, err := input.include.RelationStorage()
	if err != nil || storage != node.relationStorage {
		return RelationEdgeStatement{}, ErrContract
	}
	table, err := storage.Table()
	if err != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	binder := newBinder()
	placeholder, err := binder.bind(identifiers)
	if err != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	var sql strings.Builder
	sql.WriteString(`SELECT "source_id", "target_id", "edge_slot" FROM `)
	if appendIdentifier(&sql, table) != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	sql.WriteString(` WHERE "source_id" = ANY(CAST(`)
	sql.WriteString(placeholder)
	sql.WriteString(` AS pg_catalog.uuid[])) ORDER BY "source_id" ASC, `)
	sql.WriteString(`"edge_slot" ASC, "target_id" ASC`)
	binding, err := newIncludeStatementBinding(
		graph, node, includeStatementRoleRelationEdge,
	)
	if err != nil {
		return RelationEdgeStatement{}, ErrContract
	}
	return newRelationEdgeStatement(
		sql.String(), binder.arguments, planner.StatementClassInclude,
		node.sourceCatalog, binding,
	)
}
