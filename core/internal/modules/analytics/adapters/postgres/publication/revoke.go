package publication

import (
	"context"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (storage *store) Revoke(ctx context.Context, input publicationapp.TransitionMutation) (
	publicationapp.Snapshot, bool, error,
) {
	return storage.transition(ctx, revokeSQL, input, false)
}
