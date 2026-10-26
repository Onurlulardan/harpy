package postgres

import (
	"context"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) ListMessages(ctx context.Context, recipient domain.ID,
	filter ports.MessageFilter,
) (ports.MessagePage, error) {
	if !store.usable(ctx, false) || !recipient.Valid() || filter.Limit < 1 ||
		filter.Limit > 100 {
		return ports.MessagePage{}, ErrInvalid
	}
	before, err := parseCursor(filter.Before)
	if err != nil {
		return ports.MessagePage{}, ErrInvalid
	}
	category := ""
	if filter.Category != nil {
		category = string(*filter.Category)
	}
	rows, err := store.query(ctx, listMessagesSQL, recipient.String(),
		filter.UnreadOnly, category, before, filter.Limit+1)
	if err != nil {
		return ports.MessagePage{}, err
	}
	defer rows.Close()
	page := ports.MessagePage{Items: make([]domain.Message, 0, filter.Limit)}
	for rows.Next() {
		message, scanErr := scanMessage(rows)
		if scanErr != nil {
			return ports.MessagePage{}, scanErr
		}
		page.Items = append(page.Items, message)
	}
	if err = rows.Err(); err != nil {
		return ports.MessagePage{}, err
	}
	if len(page.Items) > filter.Limit {
		page.Items = page.Items[:filter.Limit]
		page.NextCursor = strconv.FormatInt(page.Items[len(page.Items)-1].InboxRevision, 10)
	}
	err = store.queryRow(ctx, countUnreadSQL, recipient.String()).Scan(&page.Unread)
	return page, err
}

func parseCursor(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 1 {
		return 0, ErrInvalid
	}
	return parsed, nil
}
