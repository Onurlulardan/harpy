package purge

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (command *Command) Resume(ctx context.Context, store Store,
	itemID, afterDeletionID domain.ID, limit int,
) (Execution, error) {
	if command == nil || foundation.NilPort(store) || ctx == nil || ctx.Err() != nil ||
		!itemID.UUIDv7() || afterDeletionID.Valid() && !afterDeletionID.UUIDv7() ||
		limit < 1 || limit > domain.MaximumArchiveFiles {
		return Execution{}, ErrInvalid
	}
	candidate, err := store.Lock(ctx, itemID, command.clock.Now())
	if err != nil {
		return Execution{}, err
	}
	mutation, err := buildMutation(command.ids, command.clock, candidate)
	if err != nil {
		return Execution{}, err
	}
	var result Result
	if candidate.State == domain.ItemPurging {
		result, err = store.PersistCompletion(ctx, mutation)
		if err != nil || result.State == domain.ItemPurged {
			return Execution{Result: result}, err
		}
	} else {
		result, err = store.Begin(ctx, mutation)
		if err != nil {
			return Execution{}, err
		}
	}
	page, err := store.ScanPending(ctx, itemID, afterDeletionID, uint16(limit))
	if err != nil {
		return Execution{}, err
	}
	return Execution{Result: result,
		Deletions: append([]Deletion(nil), page.Items...), More: page.More}, nil
}
