package analytics

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type analyticsCacheWarmCommand struct{ store analyticsJobStore }

func (command *analyticsCacheWarmCommand) Execute(
	ctx context.Context, claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.store == nil || ctx == nil || ctx.Err() != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"Analytics cache warm command is unavailable")
	}
	parsed, err := parseAnalyticsQueryJobClaim(claim,
		asyncmanifestv1.AnalyticsCacheWarmJobKind, "analytics-cache-warm")
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	if err := command.store.WarmQuery(
		ctx, parsed.queryID, parsed.versionID, parsed.manifest); err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	return platformasync.CompleteDatabaseJob(), nil
}
