package archive

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type ExpiryCommand struct {
	ids   foundation.IDSource
	clock foundation.Clock
}

type ExpiryExecution struct {
	Snapshot
	DeletionID domain.ID
	Created    bool
}

func NewExpiryCommand(ids foundation.IDSource,
	clock foundation.Clock,
) (*ExpiryCommand, error) {
	if foundation.NilPort(ids) || foundation.NilPort(clock) {
		return nil, fmt.Errorf("documents archive expiry command is invalid")
	}
	return &ExpiryCommand{ids: ids, clock: clock}, nil
}

func (command *ExpiryCommand) Expire(ctx context.Context, store Store,
	archiveID domain.ID,
) (ExpiryExecution, error) {
	if command == nil || foundation.NilPort(store) || ctx == nil || ctx.Err() != nil ||
		!archiveID.UUIDv7() {
		return ExpiryExecution{}, ErrInvalid
	}
	deletionID, err := command.ids.NewUUIDv7()
	if err != nil || !deletionID.UUIDv7() {
		return ExpiryExecution{}, ErrUnavailable
	}
	snapshot, created, err := store.Expire(ctx, archiveID, deletionID,
		command.clock.Now())
	if err != nil {
		return ExpiryExecution{}, err
	}
	return ExpiryExecution{Snapshot: snapshot, DeletionID: deletionID,
		Created: created}, nil
}
