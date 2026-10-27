package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func (compiler Compiler) CompileIncludeTarget(
	plan resolved.AuthorizedPlan,
	graph IncludeGraph,
	node IncludeNode,
	candidateIDs []uuid.UUID,
) (IncludeTargetStatement, error) {
	if compiler.Validate() != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	input, err := includeInputForPlan(plan, graph, node)
	if err != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	identifiers, err := normalizeIncludeUUIDs(
		candidateIDs, maximumIncludeCandidateIDs,
	)
	if err != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	columns, stableIndex, err := includeTargetColumns(input.include)
	if err != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	storage, storageErr := input.include.TargetStorage()
	target, targetErr := input.include.Target()
	stable, stableErr := input.include.StableID()
	stableID, stableIDErr := stable.ID()
	if storageErr != nil || targetErr != nil || stableErr != nil || stableIDErr != nil ||
		storage != node.targetStorage || target != node.targetCatalog {
		return IncludeTargetStatement{}, ErrContract
	}
	table, err := storage.Table()
	if err != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	binder := newBinder()
	candidates, err := binder.bind(identifiers)
	if err != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	conditions, err := includeTargetConditions(plan, input, candidates, binder)
	if err != nil {
		return IncludeTargetStatement{}, err
	}
	var sql strings.Builder
	sql.WriteString("SELECT ")
	if appendProjectionList(&sql, columns) != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	sql.WriteString(" FROM ")
	if appendIdentifier(&sql, table) != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	sql.WriteString(" WHERE ")
	sql.WriteString(strings.Join(conditions, " AND "))
	sorts, err := input.include.Sorts()
	if err != nil || appendIncludeTargetOrder(&sql, sorts, binder) != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	binding, err := newIncludeStatementBinding(
		graph, node, includeStatementRoleTargetRecord,
	)
	if err != nil {
		return IncludeTargetStatement{}, ErrContract
	}
	return newIncludeTargetStatement(
		sql.String(), binder.arguments, resultColumns(columns), stableIndex,
		stableID, planner.StatementClassInclude, target, binding,
	)
}
