package analytics

import (
	"context"
	"fmt"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type analyticsQueryRevalidateCommand struct{ store analyticsJobStore }

func (command *analyticsQueryRevalidateCommand) Execute(
	ctx context.Context, claim platformasync.DatabaseJobClaim,
) (platformasync.DatabaseJobOutcome, error) {
	if command == nil || command.store == nil || ctx == nil || ctx.Err() != nil {
		return platformasync.DatabaseJobOutcome{}, fmt.Errorf(
			"Analytics revalidation command is unavailable")
	}
	parsed, err := parseAnalyticsQueryJobClaim(claim,
		asyncmanifestv1.AnalyticsQueryRevalidateJobKind, "analytics-query-revalidate")
	if err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	if err := command.store.VerifyRevalidated(
		ctx, parsed.queryID, parsed.versionID, parsed.manifest); err != nil {
		return platformasync.DatabaseJobOutcome{}, err
	}
	return platformasync.CompleteDatabaseJob(), nil
}
