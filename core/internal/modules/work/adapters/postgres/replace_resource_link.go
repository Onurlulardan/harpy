package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/release_task_resource_links.sql
var releaseTaskResourceLinksSQL string

func (store *Store) ReplaceResourceLinks(
	ctx context.Context, taskID domain.ID, replacements []domain.ResourceLink,
) error {
	if !taskID.Valid() || len(replacements) > 2 {
		return application.ErrInvalid
	}
	for _, replacement := range replacements {
		if !replacement.Valid() || replacement.WorkResourceID != taskID {
			return application.ErrInvalid
		}
	}
	if err := store.check(ctx); err != nil {
		return err
	}
	if _, err := store.database.run(ctx, releaseTaskResourceLinksSQL, taskID); err != nil {
		return err
	}
	for _, replacement := range replacements {
		if err := store.InsertResourceLink(ctx, replacement); err != nil {
			return err
		}
	}
	return nil
}
