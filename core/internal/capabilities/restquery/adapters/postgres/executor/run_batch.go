package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func runBatch(
	ctx context.Context,
	driver execution.QueryDriver,
	batch consumedBatch,
	codec cursor.Codec,
	policy planner.Policy,
) (Result, error) {
	if ctx == nil || ctx.Err() != nil || interfaceIsNil(driver) ||
		codec.Validate() != nil || validateDriverOwner(driver, batch) != nil {
		return Result{}, ErrContract
	}
	statements, err := statementsFor(batch)
	if err != nil {
		return Result{}, err
	}
	if err := admitStatements(ctx, driver, policy, statements); err != nil {
		return Result{}, err
	}
	page, err := readRecords(
		ctx, driver, policy, batch.record, statements[0], codec,
	)
	if err != nil {
		return Result{}, err
	}
	var total int64
	if batch.hasExactCount {
		total, err = readExactCount(ctx, driver, policy, statements[1])
		if err != nil {
			return Result{}, err
		}
	}
	return newResult(page, total, batch.hasExactCount)
}

func validateDriverOwner(driver execution.QueryDriver, batch consumedBatch) error {
	if batch.Validate() != nil {
		return ErrContract
	}
	owner, err := batch.record.Owner()
	if err != nil || driver.Owner() != owner {
		return ErrContract
	}
	return nil
}
