package executor

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"

const (
	maximumGraphInteractiveStatements = 7
	maximumGraphExactStatements       = 1
)

type graphStatementBound struct {
	interactive int
	exact       int
}

func (bound graphStatementBound) add(
	submissions []graphStageSubmission,
) (graphStatementBound, error) {
	next := bound
	for _, submission := range submissions {
		if submission.waiting {
			continue
		}
		for _, candidate := range submission.work.statements {
			switch candidate.class {
			case planner.StatementClassRecord, planner.StatementClassInclude:
				next.interactive++
			case planner.StatementClassExactCount:
				next.exact++
			default:
				return bound, ErrContract
			}
		}
	}
	if next.interactive > maximumGraphInteractiveStatements ||
		next.exact > maximumGraphExactStatements {
		return bound, ErrContract
	}
	return next, nil
}
