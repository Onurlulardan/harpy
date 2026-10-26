package notificationaccess

import (
	"context"
	"time"

	notificationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/notificationaccess"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type boundReader struct {
	database platformpostgres.BoundExecutor
}

type reader struct {
	binder platformpostgres.OwnerBinder[*boundReader]
}

func NewReader() (notificationapp.SessionReader, error) {
	binder, err := platformpostgres.NewOwnerBinder("iam",
		func(database platformpostgres.BoundExecutor) *boundReader {
			return &boundReader{database: database}
		})
	if err != nil {
		return nil, notificationapp.ErrSessionUnavailable
	}
	return &reader{binder: binder}, nil
}

func (reader *reader) sessionSelector(ctx context.Context,
	scope platformtx.TenantScope, principalID, sessionID uuid.UUID,
	now time.Time,
) ([16]byte, error) {
	if reader == nil || ctx == nil || ctx.Err() != nil || principalID == uuid.Nil ||
		sessionID == uuid.Nil || now.IsZero() || platformtx.NetworkAllowed(ctx) ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.SnapshotQuery) != nil ||
		platformtx.RequireOwner(scope, "iam") != nil {
		return [16]byte{}, notificationapp.ErrSessionUnavailable
	}
	bound, err := platformtx.Bind(scope, reader.binder)
	if err != nil || bound == nil {
		return [16]byte{}, notificationapp.ErrSessionUnavailable
	}
	var raw []byte
	if err = bound.database.QueryRow(ctx, readSessionSelectorSQL,
		principalID, sessionID, now.UTC()).Scan(&raw); err != nil || len(raw) != 16 {
		clear(raw)
		return [16]byte{}, notificationapp.ErrSessionUnavailable
	}
	var selector [16]byte
	copy(selector[:], raw)
	clear(raw)
	if selector == [16]byte{} {
		return [16]byte{}, notificationapp.ErrSessionUnavailable
	}
	return selector, nil
}
