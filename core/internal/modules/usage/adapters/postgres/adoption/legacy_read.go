package adoption

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	app "github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

func (source *LegacySource) inspect(ctx context.Context,
	database LegacyDatabase,
) (app.Manifest, error) {
	var oid int64
	var count int64
	var root []byte
	locator := source.schema + "." + source.relation
	if err := database.QueryRow(ctx, source.manifestSQL, locator, source.defaultTenant).
		Scan(&oid, &count, &root); err != nil || oid < 1 || oid > 1<<32-1 ||
		count < 0 || len(root) != 32 {
		return app.Manifest{}, domain.ErrUnavailable
	}
	manifest := app.Manifest{Kind: app.LegacyTrace, Schema: source.schema,
		Relation: source.relation, RelationOID: uint32(oid), RowCount: count,
		SnapshotID: "usage-legacy-" + hex.EncodeToString(root)}
	copy(manifest.RootSHA256[:], root)
	if manifest.Validate() != nil {
		return app.Manifest{}, domain.ErrUnavailable
	}
	return manifest, nil
}

func (source *LegacySource) Page(ctx context.Context, snapshotID string,
	cursor *uuid.UUID, limit int,
) (app.SourcePage, error) {
	if source == nil || source.run == nil || ctx == nil || ctx.Err() != nil ||
		limit < 1 || limit > 100 {
		return app.SourcePage{}, domain.ErrInvalid
	}
	var page app.SourcePage
	err := source.run(ctx, source.schema, func(database LegacyDatabase) error {
		manifest, err := source.inspect(ctx, database)
		if err != nil || manifest.SnapshotID != snapshotID {
			return domain.ErrSourceDrift
		}
		rows, err := database.Query(ctx, source.pageSQL, source.defaultTenant,
			nullableUUID(cursor), limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		records := make([]app.LegacyRecord, 0, limit)
		for rows.Next() {
			var record app.LegacyRecord
			var durationMillis int64
			if err = rows.Scan(&record.SourceID, &record.TenantID, &record.ActorID,
				&record.Method, &record.Path, &record.StatusCode, &durationMillis,
				&record.OccurredAt); err != nil || durationMillis > 86400000 {
				return fmt.Errorf("usage legacy row is invalid")
			}
			record.EnvironmentID = source.environmentID
			record.DurationMicros = durationMillis * 1000
			record.OccurredAt = record.OccurredAt.UTC().Truncate(time.Microsecond)
			records = append(records, record)
		}
		if err = rows.Err(); err != nil {
			return err
		}
		if len(records) == 0 {
			return domain.ErrSourceDrift
		}
		after, inspectErr := source.inspect(ctx, database)
		if inspectErr != nil || !after.Same(manifest) {
			return domain.ErrSourceDrift
		}
		page, err = app.NewSourcePage(records)
		return err
	})
	if err != nil {
		if errors.Is(err, domain.ErrSourceDrift) {
			return app.SourcePage{}, domain.ErrSourceDrift
		}
		return app.SourcePage{}, domain.ErrUnavailable
	}
	return page, nil
}
