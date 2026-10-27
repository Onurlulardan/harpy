package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
)

func (runtime *graphRuntime) runNodeProgram(
	ctx context.Context,
	driver execution.SnapshotDriver,
	nodeIndex int,
) error {
	if runtime == nil || nodeIndex < 0 || nodeIndex >= len(runtime.nodes) {
		return ErrContract
	}
	node := runtime.nodes[nodeIndex]
	dependency, targetStage := 1, 2
	if node.material.depth == 2 {
		dependency, targetStage = 3, 4
	}
	if err := runtime.barrier.await(ctx, dependency); err != nil {
		return err
	}
	sourceIDs, err := runtime.runTarget(ctx, driver, nodeIndex, targetStage)
	if err != nil {
		return err
	}
	if node.material.depth == 1 && len(runtime.children(nodeIndex)) > 0 {
		return runtime.runOutgoing(
			ctx, driver, 3, nodeIndex, node.edgeOrdinal, sourceIDs,
		)
	}
	return nil
}
