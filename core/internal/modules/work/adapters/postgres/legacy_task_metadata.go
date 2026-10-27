package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/load_legacy_task_metadata_for_update.sql
var loadLegacyTaskMetadataForUpdateSQL string

//go:embed sql/tenant/upsert_legacy_task_metadata.sql
var upsertLegacyTaskMetadataSQL string

func (store *Store) LoadLegacyTaskMetadataForUpdate(
	ctx context.Context, taskID domain.ID,
) (domain.LegacyTaskMetadata, error) {
	if !taskID.Valid() {
		return domain.LegacyTaskMetadata{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.LegacyTaskMetadata{}, err
	}
	var value domain.LegacyTaskMetadata
	var end *time.Time
	var filter []byte
	err := store.database.QueryRow(ctx, loadLegacyTaskMetadataForUpdateSQL, taskID).Scan(
		&value.IsRecurring, &value.RecurrenceRule, &end, &filter)
	if err != nil {
		return domain.LegacyTaskMetadata{}, err
	}
	if end != nil {
		utc := end.UTC()
		value.RecurrenceEnd = &utc
	}
	value.FormFilter = append(json.RawMessage(nil), filter...)
	normalized, err := domain.NormalizeLegacyTaskMetadata(value)
	if err != nil {
		return domain.LegacyTaskMetadata{}, application.ErrUnavailable
	}
	return normalized, nil
}

func (store *Store) UpsertLegacyTaskMetadata(
	ctx context.Context, taskID domain.ID, value domain.LegacyTaskMetadata,
) error {
	if !taskID.Valid() {
		return application.ErrInvalid
	}
	normalized, err := domain.NormalizeLegacyTaskMetadata(value)
	if err != nil {
		return application.ErrInvalid
	}
	return executeOne(ctx, store, upsertLegacyTaskMetadataSQL, taskID,
		normalized.IsRecurring, normalized.RecurrenceRule,
		normalized.RecurrenceEnd, []byte(normalized.FormFilter))
}
