package v1adoption

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
	"github.com/google/uuid"
)

func (store *Store) LoadSourcePage(ctx context.Context, after *uuid.UUID,
	limit int,
) (app.SourcePage, error) {
	if !store.usable(ctx) || limit < 1 || limit > app.MaximumPageSize {
		return app.SourcePage{}, app.ErrInvalid
	}
	var cursor any
	if after != nil {
		cursor = *after
	}
	rows, err := store.query(ctx, readSourcePageSQL, cursor, limit+1)
	if err != nil {
		return app.SourcePage{}, app.ErrUnavailable
	}
	defer rows.Close()
	records := make([]app.SourceRecord, 0, limit)
	final := true
	for rows.Next() {
		if len(records) == limit {
			final = false
			break
		}
		value, scanErr := scanSourceRecord(rows)
		if scanErr != nil {
			return app.SourcePage{}, app.ErrUnavailable
		}
		records = append(records, value)
	}
	if err = rows.Err(); err != nil {
		return app.SourcePage{}, app.ErrUnavailable
	}
	return app.SourcePage{Records: records, Final: final}, nil
}
