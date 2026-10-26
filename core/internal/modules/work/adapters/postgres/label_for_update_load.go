package postgres

import (
	"context"
	_ "embed"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_label_for_update.sql
var getLabelForUpdateSQL string

func (store *Store) LoadLabelForUpdate(ctx context.Context, id domain.ID) (domain.Label, error) {
	if !id.Valid() {
		return domain.Label{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.Label{}, err
	}
	var labelID, boardID uuid.UUID
	var name, color string
	var version int64
	var deleted *time.Time
	err := store.database.QueryRow(ctx, getLabelForUpdateSQL, id).Scan(
		&labelID, &boardID, &name, &color, &version, &deleted)
	if err != nil {
		return domain.Label{}, err
	}
	label := domain.Label{ID: domain.ID(labelID), BoardID: domain.ID(boardID),
		Name: name, Color: color, RowVersion: version, DeletedAt: deleted}
	if label.DeletedAt != nil {
		value := label.DeletedAt.UTC()
		label.DeletedAt = &value
	}
	if !label.Valid() {
		return domain.Label{}, application.ErrUnavailable
	}
	return label, nil
}
