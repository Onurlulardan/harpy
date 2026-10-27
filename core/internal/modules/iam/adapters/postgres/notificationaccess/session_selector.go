package notificationaccess

import (
	"context"
	"time"

	notificationapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/notificationaccess"
	"github.com/google/uuid"
)

func (reader *reader) SessionSelector(ctx context.Context,
	scope notificationapp.Scope, principalID, sessionID uuid.UUID,
	now time.Time,
) ([16]byte, error) {
	return reader.sessionSelector(ctx, scope, principalID, sessionID, now)
}

var _ notificationapp.SessionReader = (*reader)(nil)
