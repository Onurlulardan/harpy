package publication

import (
	"context"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
)

func (storage *store) Approve(ctx context.Context, input publicationapp.TransitionMutation) (
	publicationapp.Snapshot, bool, error,
) {
	return storage.transition(ctx, approveSQL, input, true)
}
