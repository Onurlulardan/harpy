package jobs

import (
	"context"
	"time"

	jobapp "github.com/Onurlulardan/harpy/core/internal/modules/work/application/jobs"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

func (store *Store) LoadDueReminders(ctx context.Context, through time.Time,
	limit int,
) ([]jobapp.Reminder, error) {
	if !store.valid() || through.IsZero() || limit < 1 || limit > 10 {
		return nil, ports.ErrInvalid
	}
	values, err := store.database.query(ctx, loadDueRemindersSQL, through, limit)
	if err != nil {
		return nil, ports.ErrUnavailable
	}
	defer values.Close()
	result := make([]jobapp.Reminder, 0, limit)
	for values.Next() {
		var id uuid.UUID
		var reminder jobapp.Reminder
		if values.Scan(&id, &reminder.TaskVersion, &reminder.DueAt,
			&reminder.Title) != nil {
			return nil, ports.ErrUnavailable
		}
		reminder.TaskID = domain.ID(id)
		result = append(result, reminder)
	}
	if values.Err() != nil {
		return nil, ports.ErrUnavailable
	}
	return result, nil
}

func (store *Store) PersistReminder(ctx context.Context, reminder jobapp.Reminder,
	dispatchedAt time.Time,
) (bool, error) {
	if !store.valid() || !reminder.TaskID.Valid() || reminder.TaskVersion < 1 ||
		reminder.DueAt.IsZero() || !reminder.EventID.Valid() || dispatchedAt.IsZero() {
		return false, ports.ErrInvalid
	}
	affected, err := store.database.execute(ctx, recordReminderSQL, reminder.TaskID,
		reminder.TaskVersion, reminder.DueAt, reminder.EventID, dispatchedAt)
	if err != nil || affected < 0 || affected > 1 {
		return false, ports.ErrUnavailable
	}
	return affected == 1, nil
}
