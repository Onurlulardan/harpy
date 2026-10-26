package usage

import (
	"context"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func (commands *Commands) Complete(
	ctx context.Context,
	request TransitionRequest,
) (ReservationResult, error) {
	return commands.resolve(ctx, request, usagedomain.Committed, true)
}
