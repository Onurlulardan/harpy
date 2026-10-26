package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/google/uuid"
)

func (runtime *graphRuntime) runTarget(
	ctx context.Context,
	driver execution.SnapshotDriver,
	nodeIndex int,
	stage int,
) ([]uuid.UUID, error) {
	if runtime == nil || nodeIndex < 0 || nodeIndex >= len(runtime.nodes) {
		return nil, ErrContract
	}
	node := &runtime.nodes[nodeIndex]
	owner, err := node.compiled.Owner()
	if err != nil || owner != driver.Owner() {
		return nil, ErrContract
	}
	candidates := append([]uuid.UUID(nil), node.candidates...)
	if len(candidates) == 0 {
		node.material.records = []materialRecord{}
		if err := runtime.barrier.wait(ctx, stage, node.stageOrdinal); err != nil {
			return nil, err
		}
	} else {
		compiled, err := runtime.compiler.CompileIncludeTarget(
			runtime.plan, runtime.graph, node.compiled, candidates,
		)
		if err != nil {
			return nil, ErrContract
		}
		candidate, sealed, err := consumeIncludeTarget(
			compiled, runtime.plan, runtime.graph, node.compiled,
		)
		if err != nil {
			return nil, ErrContract
		}
		work := graphStageWork{
			stage: stage, ordinal: node.stageOrdinal, driver: driver,
			statements: []statement{candidate},
			run: func(runContext context.Context, current execution.SnapshotDriver) error {
				records, err := readIncludeTarget(
					runContext, current, runtime.policy, sealed, candidate, len(candidates),
				)
				if err == nil {
					node.material.records = records
				}
				return err
			},
		}
		if err := runtime.barrier.submit(ctx, work); err != nil {
			return nil, err
		}
	}
	if node.material.depth == 1 {
		return runtime.selectedDepthOneSources(nodeIndex)
	}
	return materialRecordIDs(node.material.records)
}
