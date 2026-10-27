package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/google/uuid"
)

type preparedRelationEdge struct {
	nodeIndex int
	statement statement
}

func (runtime *graphRuntime) runOutgoing(
	ctx context.Context,
	driver execution.SnapshotDriver,
	stage int,
	parent int,
	ordinal int,
	sourceIDs []uuid.UUID,
) error {
	children := runtime.children(parent)
	if len(children) < 1 || len(children) > 3 {
		return ErrContract
	}
	if len(sourceIDs) == 0 {
		runtime.clearOutgoing(parent)
		return runtime.barrier.wait(ctx, stage, ordinal)
	}
	prepared := make([]preparedRelationEdge, len(children))
	statements := make([]statement, len(children))
	for index, child := range children {
		node := runtime.nodes[child].compiled
		owner, err := node.SourceOwner()
		if err != nil || owner != driver.Owner() {
			return ErrContract
		}
		compiled, err := runtime.compiler.CompileRelationEdge(
			runtime.plan, runtime.graph, node, sourceIDs,
		)
		if err != nil {
			return ErrContract
		}
		statements[index], err = consumeRelationEdge(
			compiled, runtime.plan, runtime.graph, node,
		)
		if err != nil {
			return ErrContract
		}
		prepared[index] = preparedRelationEdge{
			nodeIndex: child, statement: statements[index],
		}
	}
	work := graphStageWork{
		stage: stage, ordinal: ordinal, driver: driver, statements: statements,
		run: func(runContext context.Context, current execution.SnapshotDriver) error {
			return runtime.readOutgoing(runContext, current, sourceIDs, prepared)
		},
	}
	return runtime.barrier.submit(ctx, work)
}

func (runtime *graphRuntime) readOutgoing(
	ctx context.Context,
	driver execution.SnapshotDriver,
	sourceIDs []uuid.UUID,
	prepared []preparedRelationEdge,
) error {
	for _, edge := range prepared {
		rows, err := readRelationEdges(
			ctx, driver, runtime.policy, edge.statement, len(sourceIDs),
		)
		if err != nil {
			return err
		}
		associations, candidates, err := buildAssociations(sourceIDs, rows)
		if err != nil {
			return err
		}
		runtime.nodes[edge.nodeIndex].material.associations = associations
		runtime.nodes[edge.nodeIndex].candidates = candidates
	}
	return nil
}
