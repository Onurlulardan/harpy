package postgres

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/google/uuid"
)

func (store *Store) ContactStatsRead(ctx context.Context, access application.Access,
	observedAt time.Time,
) (domain.ContactStats, error) {
	if !access.Allows(application.ScopePlatform, application.PermissionPlatformContactRead,
		application.OperationContactStats) {
		return domain.ContactStats{}, application.ErrForbidden
	}
	var stats domain.ContactStats
	err := store.within(ctx, SnapshotTransaction, func(database Database) error {
		var document []byte
		if err := database.QueryRow(ctx, contactStatsSQL, uuid.UUID(access.ActorID()),
			access.AuthorizationVersion(), access.AuthorityClock(), observedAt).
			Scan(&document); err != nil {
			return err
		}
		var err error
		stats, err = decodeContactStats(document)
		return err
	})
	return stats, store.normalize(err)
}
