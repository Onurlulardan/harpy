package postgres

import (
	"context"
	_ "embed"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/insert_generated_event.sql
var insertGeneratedEventSQL string

func (store *Store) InsertGeneratedEvents(ctx context.Context, ruleID domain.ID, values []ports.GeneratedOccurrence) (int, error) {
	if !ruleID.Valid() || len(values) > domain.MaxFuturePerRule {
		return 0, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return 0, err
	}
	inserted := int64(0)
	for _, value := range values {
		occurrence := value.Occurrence
		if !value.ID.Valid() || occurrence.Start.IsZero() || occurrence.TZDBVersion == "" || len(occurrence.LocalLabel) < 10 {
			return 0, application.ErrInvalid
		}
		result, err := store.database.run(ctx, insertGeneratedEventSQL, ruleID, value.ID,
			occurrence.Start.UTC(), occurrence.TZDBVersion, occurrence.LocalLabel[:10])
		if err != nil {
			return 0, err
		}
		inserted += result.RowsAffected()
	}
	if inserted < 0 || inserted > int64(len(values)) {
		return 0, application.ErrUnavailable
	}
	return int(inserted), nil
}

var _ = strings.TrimSpace
