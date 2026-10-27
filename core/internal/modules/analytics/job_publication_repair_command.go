package analytics

import (
	"context"
	"fmt"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type analyticsPublicationRepairCommand struct {
	store analyticsJobStore
	clock func() time.Time
}

func (command *analyticsPublicationRepairCommand) Execute(
	ctx context.Context, claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.store == nil || command.clock == nil ||
		ctx == nil || ctx.Err() != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"Analytics publication repair command is unavailable")
	}
	parsed, err := parseAnalyticsPublicationJobClaim(
		claim, asyncmanifestv1.AnalyticsPublicationRepairJobKind)
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	now := command.clock().UTC().Truncate(time.Microsecond)
	if now.IsZero() {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"Analytics publication repair clock is invalid")
	}
	if err := command.store.ApplyPublicationRepair(
		ctx, parsed.releaseID, parsed.generation, now); err != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"repair Analytics publication %s: %w", parsed.state, err)
	}
	return platformasync.CompleteDatabaseJob(), nil
}
