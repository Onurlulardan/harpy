package work

import (
	"context"
	"time"

	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type workDatabaseJobCommand struct {
	runner   *jobapp.Runner
	buildErr error
	now      func() time.Time
}

var _ platformasync.DatabaseJobCommand = (*workDatabaseJobCommand)(nil)

func (command *workDatabaseJobCommand) Execute(ctx context.Context,
	claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.runner == nil || command.buildErr != nil || command.now == nil {
		return platformasync.DatabaseJobOutcome{}, ports.ErrUnavailable
	}
	now := command.now().UTC().Truncate(time.Microsecond)
	parsed, err := parseWorkJobClaim(claim, now)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	result, err := command.runner.RunJob(ctx, parsed)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	if result.Duplicate || len(result.Reminders) == 0 {
		return platformasync.CompleteDatabaseJob(), nil
	}
	events, err := workReminderEvents(parsed, result.Reminders)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	return platformasync.CompleteDatabaseJobWithEffects(
		platformasync.DatabaseJobEffects{Events: events},
	)
}
