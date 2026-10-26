package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
)

func (runtime *graphRuntime) runRootProgram(
	ctx context.Context,
	driver execution.SnapshotDriver,
) error {
	statements, err := statementsFor(runtime.batch)
	if err != nil || len(statements) < 1 || len(statements) > 2 {
		return ErrContract
	}
	work := graphStageWork{
		stage: 0, ordinal: 0, driver: driver, statements: statements,
		run: func(runContext context.Context, current execution.SnapshotDriver) error {
			page, records, err := readGraphRoot(
				runContext, current, runtime.policy, runtime.batch.record,
				statements[0], runtime.codec, runtime.rootIDAt,
			)
			if err != nil {
				return err
			}
			var total int64
			if runtime.batch.hasExactCount {
				total, err = readExactCount(
					runContext, current, runtime.policy, statements[1],
				)
				if err != nil {
					return err
				}
			}
			runtime.page, runtime.root, runtime.total = page, records, total
			return nil
		},
	}
	if err := runtime.barrier.submit(ctx, work); err != nil {
		return err
	}
	sourceIDs, err := materialRecordIDs(runtime.root)
	if err != nil {
		return err
	}
	return runtime.runOutgoing(ctx, driver, 1, -1, 0, sourceIDs)
}
