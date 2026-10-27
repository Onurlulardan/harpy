package jobs

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

const reminderBatchLimit = 10

func (runner *Runner) dispatchReminders(ctx context.Context, claim Claim) (Result, error) {
	reminders, err := runner.store.LoadDueReminders(ctx, claim.Now, reminderBatchLimit)
	if err != nil || len(reminders) > reminderBatchLimit {
		return Result{}, ports.ErrUnavailable
	}
	result := make([]Reminder, 0, len(reminders))
	for _, reminder := range reminders {
		if !reminder.TaskID.Valid() || reminder.TaskVersion < 1 || reminder.DueAt.IsZero() ||
			strings.TrimSpace(reminder.Title) == "" || len(reminder.Title) > 1200 {
			return Result{}, ports.ErrInvalid
		}
		reminder.EventID, err = runner.ids.New(claim.Now)
		if err != nil || !reminder.EventID.UUIDv7() {
			return Result{}, ports.ErrUnavailable
		}
		inserted, recordErr := runner.store.PersistReminder(ctx, reminder, claim.Now)
		if recordErr != nil {
			return Result{}, recordErr
		}
		if inserted {
			result = append(result, reminder)
		}
	}
	return Result{Reminders: result}, nil
}
