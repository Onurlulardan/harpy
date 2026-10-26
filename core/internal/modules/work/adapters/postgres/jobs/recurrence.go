package jobs

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (store *Store) LoadRecurrence(ctx context.Context,
	ruleID domain.ID,
) (ports.RecurrenceRevision, error) {
	if !store.valid() || !ruleID.Valid() {
		return ports.RecurrenceRevision{}, ports.ErrInvalid
	}
	var result ports.RecurrenceRevision
	if store.database.queryRow(ctx, loadRecurrenceSQL, ruleID).Scan(&result.Revision,
		&result.Contract, &result.DTSTART, &result.TimeZone, &result.TZDBVersion,
		&result.RRULE, &result.EXDATE) != nil {
		return ports.RecurrenceRevision{}, ports.ErrUnavailable
	}
	return result, nil
}

func (store *Store) FutureOccurrenceCount(ctx context.Context) (int, error) {
	var count int
	if !store.valid() || store.database.queryRow(ctx, countFutureSQL).Scan(&count) != nil || count < 0 {
		return 0, ports.ErrUnavailable
	}
	return count, nil
}

func (store *Store) InsertOccurrences(ctx context.Context, ruleID domain.ID,
	values []ports.GeneratedOccurrence,
) (int, error) {
	if !store.valid() || !ruleID.Valid() || len(values) > domain.MaxFuturePerRule {
		return 0, ports.ErrInvalid
	}
	inserted := int64(0)
	for _, value := range values {
		occurrence := value.Occurrence
		if !value.ID.Valid() || occurrence.Start.IsZero() ||
			occurrence.TZDBVersion == "" || len(occurrence.LocalLabel) < 10 {
			return 0, ports.ErrInvalid
		}
		affected, err := store.database.execute(ctx, insertOccurrenceSQL, ruleID,
			value.ID, occurrence.Start, occurrence.TZDBVersion, occurrence.LocalLabel[:10])
		if err != nil {
			return 0, ports.ErrUnavailable
		}
		inserted += affected
	}
	return int(inserted), nil
}
