package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func runV1(
	ctx context.Context,
	driver execution.SnapshotDriver,
	plan resolved.AuthorizedPlan,
	ordinal v1query.PredecessorOrdinal,
	predecessor compiler.PredecessorStatement,
	batch consumedBatch,
	postgresCompiler compiler.Compiler,
	codec cursor.Codec,
	policy planner.Policy,
) (Result, error) {
	if ctx == nil || ctx.Err() != nil || interfaceIsNil(driver) ||
		batch.Validate() != nil || !batch.hasExactCount ||
		postgresCompiler.Validate() != nil || codec.Validate() != nil ||
		predecessor.ValidatePlan(plan, ordinal) != nil ||
		validateV1DriverOwner(driver, batch, predecessor) != nil {
		return Result{}, ErrContract
	}
	seek, err := predecessorExecutorStatement(predecessor)
	if err != nil {
		return Result{}, err
	}
	count, err := countStatement(batch.exactCount)
	if err != nil {
		return Result{}, err
	}
	tracker, err := planner.NewTracker(policy)
	if err != nil {
		return Result{}, ErrContract
	}
	tracker, err = admitStage(ctx, driver, policy, tracker, []statement{seek, count})
	if err != nil {
		return Result{}, err
	}
	tuple, found, err := readV1Predecessor(ctx, driver, policy, predecessor, seek)
	if err != nil {
		return Result{}, err
	}
	page, err := runV1Page(
		ctx, driver, plan, ordinal, predecessor, tuple, found, batch,
		postgresCompiler, codec, policy, &tracker,
	)
	if err != nil {
		return Result{}, err
	}
	total, err := readExactCount(ctx, driver, policy, count)
	if err != nil {
		return Result{}, err
	}
	return newResult(page, total, true)
}

func runV1Page(
	ctx context.Context,
	driver execution.SnapshotDriver,
	plan resolved.AuthorizedPlan,
	ordinal v1query.PredecessorOrdinal,
	predecessor compiler.PredecessorStatement,
	tuple compiler.PredecessorTuple,
	found bool,
	batch consumedBatch,
	postgresCompiler compiler.Compiler,
	codec cursor.Codec,
	policy planner.Policy,
	tracker *planner.Tracker,
) (recordpage.Page, error) {
	if !found {
		page, err := recordpage.Assemble(batch.record, codec, nil)
		if err != nil {
			return recordpage.Page{}, ErrContract
		}
		return page, nil
	}
	compiled, err := postgresCompiler.CompileV1Page(
		plan, ordinal, predecessor, tuple,
	)
	if err != nil {
		return recordpage.Page{}, ErrContract
	}
	candidate, err := recordStatement(compiled)
	if err != nil || tracker == nil {
		return recordpage.Page{}, ErrContract
	}
	*tracker, err = admitStage(
		ctx, driver, policy, *tracker, []statement{candidate},
	)
	if err != nil {
		return recordpage.Page{}, err
	}
	return readRecords(ctx, driver, policy, compiled, candidate, codec)
}

func validateV1DriverOwner(
	driver execution.SnapshotDriver,
	batch consumedBatch,
	predecessor compiler.PredecessorStatement,
) error {
	owner, err := predecessor.Owner()
	if err != nil || validateDriverOwner(driver, batch) != nil || driver.Owner() != owner {
		return ErrContract
	}
	return nil
}
