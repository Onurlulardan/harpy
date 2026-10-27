package calendar

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) ArchiveEvent(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, eventID domain.ID) (ports.OperationReceipt, error) {
	return service.writeArchived(ctx, scope, meta, eventID, true)
}

func (service *Service) writeArchived(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, eventID domain.ID, archived bool) (ports.OperationReceipt, error) {
	if !eventID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(eventID.String(), ports.VersionString(boolVersion(archived)),
		ports.VersionString(meta.ExpectedVersion))
	return service.mutator.Mutate(ctx, scope, "work.calendar.change", eventID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			event, err := stores.Calendar().LoadEventForUpdate(ctx, eventID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if event.RowVersion != meta.ExpectedVersion || (event.DeletedAt != nil) == archived {
				return ports.OperationReceipt{}, nil, ports.ErrConflict
			}
			if archived {
				now := service.clock.Now().UTC()
				event.DeletedAt = &now
			} else {
				event.DeletedAt = nil
			}
			event.RowVersion++
			if err = stores.Calendar().PersistEvent(ctx, event, meta.ExpectedVersion); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			return ports.OperationReceipt{AggregateID: eventID, RowVersion: event.RowVersion}, []byte(`{"changed":true}`), nil
		})
}

func boolVersion(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
