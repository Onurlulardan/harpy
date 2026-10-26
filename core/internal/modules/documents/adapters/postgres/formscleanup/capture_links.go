package formscleanup

import (
	"context"
	"time"

	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) CaptureLinks(ctx context.Context, operationID domain.ID,
	now time.Time,
) (formscleanupapp.Snapshot, []formscleanupapp.Candidate, error) {
	if !storage.usable(ctx) || !operationID.UUIDv7() || now.IsZero() {
		return formscleanupapp.Snapshot{}, nil, formscleanupapp.ErrInvalid
	}
	found, err := storage.query(ctx, prepareSQL, operationID.String(), now)
	if err != nil {
		return formscleanupapp.Snapshot{}, nil, formscleanupapp.ErrUnavailable
	}
	defer found.Close()
	var snapshot formscleanupapp.Snapshot
	candidates := make([]formscleanupapp.Candidate, 0)
	for found.Next() {
		var itemID, jobID dbText
		var runAt dbInstant
		next, scanErr := scanSnapshot(found, &itemID, &jobID, &runAt)
		if scanErr != nil {
			return formscleanupapp.Snapshot{}, nil, scanErr
		}
		if !snapshot.OperationID.Valid() {
			snapshot = next
		} else if snapshot != next {
			return formscleanupapp.Snapshot{}, nil, formscleanupapp.ErrUnavailable
		}
		if itemID.Valid {
			item, parseErr := domain.ParseID(itemID.Text)
			if parseErr != nil || !item.UUIDv7() || !runAt.Valid {
				return formscleanupapp.Snapshot{}, nil, formscleanupapp.ErrUnavailable
			}
			job := domain.ID{}
			if jobID.Valid {
				job, parseErr = domain.ParseID(jobID.Text)
				if parseErr != nil || !job.UUIDv7() {
					return formscleanupapp.Snapshot{}, nil, formscleanupapp.ErrUnavailable
				}
			}
			candidates = append(candidates, formscleanupapp.Candidate{
				ItemID: item, JobID: job, PurgeRunAt: runAt.Time.UTC(),
			})
		}
	}
	if err = found.Err(); err != nil || !snapshot.OperationID.Valid() {
		return formscleanupapp.Snapshot{}, nil, formscleanupapp.ErrUnavailable
	}
	return snapshot, candidates, nil
}
