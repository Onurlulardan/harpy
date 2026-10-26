package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type graphCandidate struct {
	workIndex int
	statement statement
}

func runGraphStage(
	ctx context.Context,
	policy planner.Policy,
	tracker planner.Tracker,
	submissions []graphStageSubmission,
) (planner.Tracker, error) {
	candidates := make([]graphCandidate, 0)
	for workIndex, submission := range submissions {
		if submission.waiting {
			continue
		}
		for _, candidate := range submission.work.statements {
			candidates = append(candidates, graphCandidate{
				workIndex: workIndex, statement: candidate,
			})
		}
	}
	next, err := admitGraphCandidates(ctx, policy, tracker, submissions, candidates)
	if err != nil {
		return tracker, err
	}
	for _, submission := range submissions {
		if submission.waiting {
			continue
		}
		if err := submission.work.run(ctx, submission.work.driver); err != nil {
			return tracker, err
		}
	}
	return next, nil
}

func admitGraphCandidates(
	ctx context.Context,
	policy planner.Policy,
	tracker planner.Tracker,
	submissions []graphStageSubmission,
	candidates []graphCandidate,
) (planner.Tracker, error) {
	costs := make([]planner.Cost, len(candidates))
	for index, candidate := range candidates {
		driver := submissions[candidate.workIndex].work.driver
		if err := applyTimeout(ctx, driver, policy, candidate.statement.class); err != nil {
			return tracker, err
		}
		cost, err := explainCost(ctx, driver, policy, candidate.statement)
		if err != nil {
			return tracker, err
		}
		costs[index] = cost
	}
	next := tracker
	for index, candidate := range candidates {
		var err error
		next, err = next.Admit(candidate.statement.class, costs[index])
		if err != nil {
			return tracker, err
		}
	}
	return next, nil
}
