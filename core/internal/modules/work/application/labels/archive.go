package labels

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) ArchiveLabel(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, boardID, labelID domain.ID) (ports.OperationReceipt, error) {
	return service.writeArchived(ctx, scope, meta, boardID, labelID, true)
}

func (service *Service) writeArchived(ctx context.Context, scope ports.Scope,
	meta ports.Mutation, boardID, labelID domain.ID, archived bool,
) (ports.OperationReceipt, error) {
	if !boardID.Valid() || !labelID.Valid() || meta.ExpectedVersion < 1 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	hash := ports.Fingerprint(boardID.String(), labelID.String(), ports.VersionString(meta.ExpectedVersion),
		ports.VersionString(boolVersion(archived)))
	return service.mutator.Mutate(ctx, scope, "work.label.change", boardID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			var deletedAt *time.Time
			if archived {
				now := service.clock.Now().UTC()
				deletedAt = &now
			}
			if err := stores.Tasks().PersistLabelArchive(ctx, labelID,
				meta.ExpectedVersion, deletedAt); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			return ports.OperationReceipt{AggregateID: labelID,
				RowVersion: meta.ExpectedVersion + 1}, []byte(`{"archived":true}`), nil
		})
}
