package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func admitStatements(
	ctx context.Context,
	driver execution.QueryDriver,
	policy planner.Policy,
	statements []statement,
) error {
	tracker, err := planner.NewTracker(policy)
	if err != nil || len(statements) == 0 {
		return ErrContract
	}
	_, err = admitStage(ctx, driver, policy, tracker, statements)
	return err
}

func admitStage(
	ctx context.Context,
	driver execution.QueryDriver,
	policy planner.Policy,
	tracker planner.Tracker,
	statements []statement,
) (planner.Tracker, error) {
	if len(statements) == 0 {
		return tracker, ErrContract
	}
	var err error
	costs := make([]planner.Cost, len(statements))
	for index, candidate := range statements {
		if err := applyTimeout(ctx, driver, policy, candidate.class); err != nil {
			return tracker, err
		}
		costs[index], err = explainCost(ctx, driver, policy, candidate)
		if err != nil {
			return tracker, err
		}
	}
	for index, candidate := range statements {
		tracker, err = tracker.Admit(candidate.class, costs[index])
		if err != nil {
			return tracker, err
		}
	}
	return tracker, nil
}

func explainCost(
	ctx context.Context,
	driver execution.QueryDriver,
	policy planner.Policy,
	candidate statement,
) (planner.Cost, error) {
	row := driver.QueryRow(ctx, candidate.explainSQL, candidate.arguments...)
	if interfaceIsNil(row) {
		return planner.Cost{}, ErrContract
	}
	var document []byte
	if err := row.Scan(&document); err != nil {
		return planner.Cost{}, mapScanError(ctx, policy, candidate.class, err)
	}
	cost, err := planner.ParseExplainCost(document)
	if err != nil {
		return planner.Cost{}, ErrContract
	}
	return cost, nil
}
