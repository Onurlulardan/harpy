package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func (store *Store) AcquireRealtimeHints(ctx context.Context,
	limit int,
) ([]ports.RealtimeHint, error) {
	if !store.usable(ctx, true) || limit < 1 || limit > 500 {
		return nil, ErrInvalid
	}
	rows, err := store.query(ctx, claimRealtimeHintsSQL, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]ports.RealtimeHint, 0, limit)
	for rows.Next() {
		var value ports.RealtimeHint
		var messageID, recipientID string
		if err = rows.Scan(&messageID, &recipientID, &value.InboxRevision); err != nil {
			return nil, err
		}
		value.MessageID, err = decodeID(messageID)
		if err == nil {
			value.RecipientID, err = decodeID(recipientID)
		}
		if err != nil || value.InboxRevision < 1 {
			return nil, ErrInvalid
		}
		result = append(result, value)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

var _ ports.RealtimeProjectionStore = (*Store)(nil)
