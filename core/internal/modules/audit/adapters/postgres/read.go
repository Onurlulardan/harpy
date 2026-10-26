package postgres

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
)

func (store *Store) List(
	ctx context.Context, filter domain.Filter,
) (domain.Page, error) {
	if !store.allowed(ctx, application.ScopeSnapshot) || filter.Validate() != nil {
		return domain.Page{}, application.ErrInvalid
	}
	resultRows, err := store.query(ctx, listSQL, listArguments(filter,
		filter.Page, filter.Limit)...)
	if err != nil {
		return domain.Page{}, mapError(err, store.notFound)
	}
	defer resultRows.Close()
	page := domain.Page{Entries: make([]domain.Entry, 0, filter.Limit),
		Page: filter.Page, Limit: filter.Limit}
	var total int64
	for resultRows.Next() {
		var rowTotal int64
		entry, scanErr := scanEntry(resultRows, store.notFound, &rowTotal)
		if scanErr != nil || rowTotal < 0 || total != 0 && total != rowTotal {
			return domain.Page{}, application.ErrUnavailable
		}
		total = rowTotal
		page.Entries = append(page.Entries, entry)
	}
	if err := resultRows.Err(); err != nil {
		return domain.Page{}, mapError(err, store.notFound)
	}
	if len(page.Entries) == 0 && filter.Page > 1 {
		total, err = store.firstPageTotal(ctx, filter)
		if err != nil {
			return domain.Page{}, err
		}
	}
	page.Total = int(total)
	if page.Total > 0 {
		page.TotalPages = (page.Total + page.Limit - 1) / page.Limit
	}
	return page, nil
}

func listArguments(filter domain.Filter, page, limit int) []any {
	return []any{
		optionalText(filter.OwnerModule), optionalText(filter.EntityType),
		optionalText(filter.EntityID), optionalText(filter.FormSlug),
		optionalText(string(filter.Action)), optionalUUIDArgument(filter.ActorID),
		optionalUUIDArgument(filter.OrganizationID), optionalUUIDArgument(filter.CorrelationID),
		filter.Start, filter.End, optionalText(filter.Search), page, limit,
	}
}

func (store *Store) firstPageTotal(ctx context.Context, filter domain.Filter) (int64, error) {
	var total int64
	_, err := scanEntry(store.queryRow(ctx, listSQL,
		listArguments(filter, 1, 1)...), store.notFound, &total)
	if errors.Is(err, application.ErrNotFound) {
		return 0, nil
	}
	if err != nil || total < 1 {
		return 0, application.ErrUnavailable
	}
	return total, nil
}

func (store *Store) Get(
	ctx context.Context, id domain.ID,
) (domain.Entry, error) {
	if !store.allowed(ctx, application.ScopeSnapshot) || !id.UUIDv7() {
		return domain.Entry{}, application.ErrInvalid
	}
	return scanEntry(store.queryRow(ctx, getSQL, uuidArgument(id)), store.notFound)
}
